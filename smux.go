package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"

	sio "genja.org/smux/io"
	"genja.org/smux/sink"
)

var outWriter = sio.NewMultiWriter()
var errWriter = sio.NewMultiWriter()


func execCommand(cmd string, args []string) *exec.Cmd {
	binary := exec.Command(cmd, args...)
	binary.Stdout = errWriter
	binary.Stderr = outWriter
	return binary
}

func appendErr (e error) {
	if e != nil {
		_, ouch := errWriter.WriteString(fmt.Sprintf("%s\n", e))
		if ouch != nil {
			panic(ouch)
		}
	}
}

// 1. read standard input
// 2. start program (pipe reader)
// 3. feed input to program (pipe writer)
func main() {
	buf, err := sio.ReadPass()
	if err != nil {
		log.Fatalln(err)
	}
	ptb := buf.Bytes()
	plainText := bytes.NewReader(ptb)

	for _, commands := range os.Args[1:] {
		cmd := strings.Split(commands, " ")
		comm, args := cmd[0], cmd[1:]

		// _,_ = plainText.Read(ptb)
		// _,_ = errWriter.WriteString(fmt.Sprintf("smux: running %s <<<'%s'\n", cmd, ptb))
		_,_ = errWriter.WriteString(fmt.Sprintf("smux: running %s\n", cmd))

		_,_ = plainText.Seek(0, 0)

		switch comm {

		case `ssh-add`:

			sock, err := net.Dial(`unix`, os.Getenv(`SSH_AUTH_SOCK`))
			if err != nil {
				appendErr(err)
			}

			var stdErr = sink.PassInputAgent(sock, args, plainText, errWriter)
			appendErr(stdErr)


		default:

			binary := execCommand(comm, args)
			stdErr := sink.PassInputGeneric(binary, plainText)
			appendErr(stdErr)
		}
		// _,_ = errWriter.WriteString(fmt.Sprintf("smux: done with %s\n", cmd))
	}


	defer func() {
		_,_ = os.Stderr.Write(errWriter.Bytes())
	}()
}



