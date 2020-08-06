
package io

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"sync"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

type MultiWriter interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	WriteString(s string) (n int, err error)
	Flush() error
	Bytes() []byte
	String() string
}


func ReadPass() (password *bytes.Buffer, err error) {
	var fi, _ = os.Stdin.Stat()

	var pass []byte

	isChardev := fi.Mode()&os.ModeCharDevice != 0
	isNamedPipe := fi.Mode()&os.ModeNamedPipe != 0
	if ! isChardev || isNamedPipe {
		sin := bufio.NewReader(os.Stdin)
		pass, _, err = sin.ReadLine()
	} else {
		_, _ = os.Stderr.WriteString(fmt.Sprint("passphrase: "))
		pass, err = terminal.ReadPassword(syscall.Stdin)
		if err != nil {
			return nil, err
		}
		defer fmt.Println()
	}
	return bytes.NewBuffer(pass), err
}

func NewMultiWriter() MultiWriter {
	bwr := make([]byte, 0)
	buf := bytes.NewBuffer(bwr)
	return mwrite{ mutex: sync.Mutex{}, buf: buf }
}


type mwrite struct {
	mutex  sync.Mutex
	buf    *bytes.Buffer
}

func (mw mwrite) String() string {
	return mw.buf.String()
}

func (mw mwrite) Read(p []byte) (n int, err error) {
	n = copy(p, mw.buf.Bytes())
	return n, nil
}


func (mw mwrite) Write(p []byte) (n int, err error) {
	mw.mutex.Lock()
	n, err = mw.buf.Write(p)
	mw.mutex.Unlock()
	return n, err
}

func (mw mwrite) Flush() error {
	return nil
}

func (mw mwrite) WriteString(s string) (n int, err error) {
	mw.mutex.Lock()
	n, err = mw.buf.Write([]byte(s))
	mw.mutex.Unlock()
	return n, err
}

func (mw mwrite) Close() error {
	// todo
	return nil
}

func (mw mwrite) Bytes() []byte {
	return mw.buf.Bytes()
}
