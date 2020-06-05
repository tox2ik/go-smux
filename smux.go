/*

References

- https://stackoverflow.com/questions/38094555/golang-read-os-stdin-input-but-dont-echo-it
- https://stackoverflow.com/questions/2137357/getpasswd-functionality-in-go
- https://stackoverflow.com/questions/10385551/get-exit-code-go

- keysAvailable(agent Socket, identities []string) - https://bitbucket.org/rw_grim/convey/src/default/ssh/agent.go

*/
package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
)

// 1. read standard input
// 2. start application
// 3. feed input to application
func main() {

	var err error
	var pass []byte

	_, _ = os.Stderr.WriteString(fmt.Sprint("give me all your secrets: "))
	if pass, err = readPass(); err != nil {
		log.Fatalln(err)
	}
	fmt.Println()
	for _, cmds := range os.Args[1:] {
		cmd := strings.Split(cmds, " ")
		comm, args := cmd[0], cmd[1:]
		if "ssh-add" == comm {
			err = passInputAgent(args, pass)
		} else {
			err = passInputGeneric(comm, args, pass)
		}
		if err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("%s %v => %s\n", comm, args, err))
		}
	}
}

func passInputGeneric(command string, rest []string, pass []byte) (err error) {
	cmd := exec.Command(command, rest...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	var stdin io.WriteCloser
	if stdin, err = cmd.StdinPipe(); err == nil {
		if err = cmd.Start(); err == nil {
			if _, err = stdin.Write(pass); err == nil {
				return nil
			}
		}
	}
	return err
}

func passInputAgent(keyPaths []string, pass []byte) (err error) {
	var sock net.Conn
	if sock, err = net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err != nil {
		return err
	}
	sshAgent := agent.NewClient(sock)
	var pk crypto.PrivateKey
	var errata = ""
	for _, keyPath := range keyPaths {
		var errK error
		var bytes []byte
		if bytes, errK = ioutil.ReadFile(keyPath); errK == nil {
			if pk, errK = parsePrivateKey(bytes, pass); errK == nil {
				errK = sshAgent.Add(agent.AddedKey{PrivateKey: pk})
			}
		}
		if errK != nil {
			errata += fmt.Sprintf("%s: %s\n", keyPath, errK)
		}
	}
	if errata != "" {
		return errors.New("\n" + errata)
	}
	return nil
}

func readPass() (pass []byte, err error) {
	if pass, err = terminal.ReadPassword(syscall.Stdin); err != nil {
		return nil, err
	}
	return pass, err
}

// https://github.com/golang/go/blob/dev.boringcrypto.go1.13/src/crypto/tls/tls.go
func parsePrivateKey(key []byte, pass []byte) (pk crypto.PrivateKey, err error) {

	if pk, err = ssh.ParseRawPrivateKeyWithPassphrase(key, pass); err == nil {
		return pk, nil
	}

	derB, _ := pem.Decode(key)
	der := derB.Bytes
	if pk, err = x509.ParsePKCS1PrivateKey(der); err == nil {
		return pk, nil
	}
	if pk, err = x509.ParsePKCS8PrivateKey(der); err == nil {
		switch pk.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey:
			return pk, nil
		default:
			return nil, errors.New("unknown private pk type in PKCS#8 wrapping")
		}
	}
	if pk, err = x509.ParseECPrivateKey(der); err == nil {
		return pk, nil
	}

	return nil, errors.New("invalid passphrase or bad key")
}
