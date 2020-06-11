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
	"bytes"
	"crypto"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
)

var outWriter = NewMultiWriter()
var errWriter = NewMultiWriter()

func appendErr (e error) {
	if e != nil {
		_, ouch := errWriter.WriteString(fmt.Sprintf("%s\n", e))
		if ouch != nil {
			panic(ouch)
		}
	}
}


// 1. read standard input
// 2. start application
// 3. feed input to application
func main() {

	pass, err := readPass()
	if err != nil {
		log.Fatalln(err)
	}
	for _, cmds := range os.Args[1:] {
		cmd := strings.Split(cmds, " ")
		comm, args := cmd[0], cmd[1:]
		if "ssh-add" == comm {
			appendErr(passInputAgent(args, pass))

		} else {
			appendErr(passInputGeneric(comm, args, pass))
		}
		_,_ = errWriter.WriteString(fmt.Sprintf("done with %s", cmd))
	}


	var ioErr error
	_, ioErr = os.Stderr.WriteString(string(errWriter.Bytes()))
	if ioErr != nil { panic(ioErr) }

	// hm.
	// _, ioErr = os.Stderr.WriteString(string(outWriter.Bytes()))
	// if ioErr != nil {
	// 	panic(ioErr)
	// }

}

func passInputGeneric(command string, rest []string, pass []byte) error {
	cmd := exec.Command(command, rest...)
	cmd.Stdout = errWriter
	cmd.Stderr = outWriter
	stdin, err := cmd.StdinPipe()
	if nil == err {
		err = cmd.Start()
	}
	if nil == err {
		_, err = stdin.Write(pass)
	}
	if nil == err {
		err = stdin.Close()
	}
	if nil == err {
		err = cmd.Wait()
	}
	return err
}

func passInputAgent(keyPaths []string, pass []byte) (err error) {
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
			pk, err = parsePrivateKey(data, pass)
		}
		if nil == err {
			lifeTime, _ := strconv.ParseInt(os.Getenv("SSH_ADD_LIFE"), 10, 32)
			err = sshAgent.Add(agent.AddedKey{PrivateKey: pk, Comment:keyPath, LifetimeSecs: uint32(lifeTime) })
		}
		if nil == err{
			_, err = errWriter.WriteString(fmt.Sprintf("Identity added: %s\n", keyPath))
		}
		if err != nil {
			failures++
			_, ouch := errWriter.WriteString(fmt.Sprintf("%s: %s\n", keyPath, err))
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

func readPass() (pass []byte, err error) {
	var fi, _ = os.Stdin.Stat()

	isChardev := fi.Mode()&os.ModeCharDevice != 0
	isNamedPipe := fi.Mode()&os.ModeNamedPipe != 0
	if ! isChardev || isNamedPipe {
		sin := bufio.NewReader(os.Stdin)
		pass, _, err = sin.ReadLine()
	} else {
		_, _ = os.Stderr.WriteString(fmt.Sprint("passphrase: "))
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

type MultiWriter struct {
	mutex  sync.Mutex
	writer io.StringWriter
	buffer *bytes.Buffer
}


func NewMultiWriter() *MultiWriter {
	w := bytes.NewBuffer(make([]byte, 4096))
	return &MultiWriter{
		mutex:  sync.Mutex{},
		buffer: w,
		writer: bufio.NewWriter(w),
	}
}

func (mw MultiWriter) Read(p []byte) (n int, err error) {
	if nil != mw.buffer {
		p = mw.buffer.Bytes()
		mw.buffer.Truncate(0)
		return len(p), nil
	}
	return 0, errors.New("null buffer")
}

func (mw MultiWriter) Write(p []byte) (n int, err error) {
	return mw.buffer.Write(p)
}

func (mw MultiWriter) WriteString(s string) (n int, err error) {
	mw.mutex.Lock()
	n, err = mw.writer.WriteString(s)
	mw.mutex.Unlock()
	return n, err
}

func (mw MultiWriter) Close() error {
	mw.buffer = nil
	return nil
}

func (mw MultiWriter) Bytes() []byte {
	return mw.buffer.Bytes()
}
