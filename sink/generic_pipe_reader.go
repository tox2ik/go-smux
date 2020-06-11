package sink

import (
	"io"
	"os/exec"
)

func PassInputGeneric(cmd *exec.Cmd, password io.Reader) (err error) {
	var plainTextPassword []byte
	_, err = password.Read(plainTextPassword)
	if err != nil {
		return err
	}


	stdin, err := cmd.StdinPipe()
	if nil == err {
		err = cmd.Start()
	}
	if nil == err {
		_, err = stdin.Write(plainTextPassword)
	}
	if nil == err {
		err = stdin.Close()
	}
	if nil == err {
		err = cmd.Wait()
	}
	return err
}
