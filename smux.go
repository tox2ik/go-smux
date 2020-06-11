package main

import (
	"fmt"
	"log"
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
// 3. feed input to program (pipe reader)
func main() {
	plainText, err := sio.ReadPass()
	if err != nil {
		log.Fatalln(err)
	}
	for _, commands := range os.Args[1:] {
		cmd := strings.Split(commands, " ")
		comm, args := cmd[0], cmd[1:]
		if "ssh-add" == comm {
			stdErr := sink.PassInputAgent(args, plainText, errWriter)
			appendErr(stdErr)

		} else {
			binary := execCommand(comm, args)
			stdErr := sink.PassInputGeneric(binary, plainText)
			appendErr(stdErr)
		}
		_,_ = errWriter.WriteString(fmt.Sprintf("done with %s", cmd))
	}

	defer os.Stderr.WriteString(string(errWriter.Bytes()))

	// hm.
	// _, ioErr = os.Stderr.WriteString(string(outWriter.Bytes()))
	// if ioErr != nil {
	// 	panic(ioErr)
	// }
}



