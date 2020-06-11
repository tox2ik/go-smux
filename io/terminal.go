
package io

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

type MultiWriter interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	WriteString(s string) (n int, err error)
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
	w := new(bytes.Buffer)
	return mwrite{
		mutex:  sync.Mutex{},
		buffer: w,
		writer: bufio.NewWriter(w),
	}
}



type mwrite struct {
	mutex  sync.Mutex
	writer io.StringWriter
	buffer *bytes.Buffer
}

func (mw mwrite) String() string {
	return string(mw.buffer.Bytes())
}

func (mw mwrite) Read(p []byte) (n int, err error) {
	if nil != mw.buffer {
		p = mw.buffer.Bytes()
		mw.buffer.Truncate(0)
		return len(p), nil
	}
	return 0, errors.New("null buffer")
}

func (mw mwrite) Write(p []byte) (n int, err error) {
	return mw.buffer.Write(p)
}

func (mw mwrite) WriteString(s string) (n int, err error) {
	mw.mutex.Lock()
	n, err = mw.writer.WriteString(s)
	mw.mutex.Unlock()
	return n, err
}

func (mw mwrite) Close() error {
	mw.buffer = nil
	return nil
}

func (mw mwrite) Bytes() []byte {
	return mw.buffer.Bytes()
}
