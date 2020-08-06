package sink

import (
	"io"
	"os/exec"
	"strings"
)

func PassInputString(cmd *exec.Cmd, stdinIngput string) (err error) {
	return PassInputGeneric(cmd, strings.NewReader(stdinIngput))
}

func PassInputGeneric(cmd *exec.Cmd, password io.Reader) (err error) {
	// var plainTextPassword []byte
	var plainTextPassword = make([]byte, 4096)
	var n int
	n, err = password.Read(plainTextPassword)
	if err != nil {
		return err
	}

	stdin, err := cmd.StdinPipe()
	if nil == err {
		err = cmd.Start()
	}
	if nil == err {
		clamp := plainTextPassword[0:n]
		n, err = stdin.Write(clamp)
	}
	if nil == err {
		err = stdin.Close()
	}
	if nil == err {
		err = cmd.Wait()
	}
	return err

}
