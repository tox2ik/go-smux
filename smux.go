/*

References

- https://stackoverflow.com/questions/38094555/golang-read-os-stdin-input-but-dont-echo-it
- https://stackoverflow.com/questions/2137357/getpasswd-functionality-in-go
- https://stackoverflow.com/questions/10385551/get-exit-code-go

- https://flaviocopes.com/go-shell-pipes/

- keysAvailable(agent Socket, identities []string) - https://bitbucket.org/rw_grim/convey/src/default/ssh/agent.go

- https://unix.stackexchange.com/questions/28503/how-can-i-send-stdout-to-multiple-commands

*/
package main

import (
	"bufio"
	"crypto"
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

	if pass, err = readPass(); err != nil {
		log.Fatalln(err)
	}
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
	var fi, _ = os.Stdin.Stat()

	isChardev := fi.Mode() & os.ModeCharDevice != 0
	isNamedPipe := fi.Mode() & os.ModeNamedPipe != 0
	if  ! isChardev || isNamedPipe {
		sin := bufio.NewReader(os.Stdin)
		pass, _, err = sin.ReadLine()
	} else {
		_, _ = os.Stderr.WriteString(fmt.Sprint("give me all your secrets: "))
		if pass, err = terminal.ReadPassword(syscall.Stdin); err != nil {
			return pass, err
		}
		defer fmt.Println()
	}
	return pass, err
}

func parsePrivateKey(key []byte, pass []byte) (pk crypto.PrivateKey, err error) {
	if pk, err = ssh.ParseRawPrivateKeyWithPassphrase(key, pass); err == nil {
		return pk, nil
	}
	if pk, err = ssh.ParseRawPrivateKey(key); err == nil {
		return pk, nil
	}
	return nil, errors.New("invalid passphrase or bad key")
}
