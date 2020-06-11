package sink

import (
	"crypto"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func PassInputAgent(keyPaths []string, password io.Reader, stdErr io.StringWriter) (err error) {
	var plainTextPassword []byte
	_, err = password.Read(plainTextPassword)
	if err != nil {
		return err
	}

	var sock net.Conn
	var pk crypto.PrivateKey
	sock, err = net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return err
	}
	sshAgent := agent.NewClient(sock)
	failures := 0
	for _, keyPath := range keyPaths {
		data, err := ioutil.ReadFile(keyPath)
		if nil == err {
			pk, err = parsePrivateKey(data, plainTextPassword)
		}
		if nil == err {
			lifeTime, _ := strconv.ParseInt(os.Getenv("SSH_ADD_LIFE"), 10, 32)
			err = sshAgent.Add(agent.AddedKey{PrivateKey: pk, Comment:keyPath, LifetimeSecs: uint32(lifeTime) })
		}
		if nil == err{
			_, err = stdErr.WriteString(fmt.Sprintf("Identity added: %s\n", keyPath))
		}
		if err != nil {
			failures++
			_, ouch := stdErr.WriteString(fmt.Sprintf("%s: %s\n", keyPath, err))
			if ouch != nil {
				log.Fatal(ouch)
			}
		}
	}

	if failures > 0 {
		return fmt.Errorf("failed to add %d keys", failures)
	}
	return nil
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

