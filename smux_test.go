package main

import (
	"bytes"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/smartystreets/assertions/should"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/ssh/agent"

	"genja.org/smux/io"
	"genja.org/smux/sink"
)

// echo -n 12345 | go run smux.go "ssh-add ./testdata/jwf.ossh ./test/fixture/jwf.ec19 ./test/fixture/jwf.clear"

const socket = `testdata/keyring.sock`

var protectedKeys = []string{
	`testdata/jwf.ossh`,
	`testdata/jwf.pem`,
	`testdata/jwf.ec19`,
}

func TestSshAgent_ImportProtectedKey(t *testing.T) {

	stdErr := io.NewMultiWriter()
	pass12345 := bytes.NewReader([]byte(`12345`))
	var keyring *exec.Cmd

	_ = exec.Command(`make`, `keys`).Run()

	Convey("Start an agent", t, func() {

		started := make(chan interface{}, 1)
		timeout := time.After(777*time.Millisecond)

		Convey("Run agent", func(c C) {
			_ = os.Remove(socket)
			keyring = exec.Command(`ssh-agent`, `-a`, socket, `-Emd5`, `-D`)
			keyring.Stderr = stdErr
			err := keyring.Start()
			So(err, ShouldBeNil)
			go startAndWait(started)
		})
		select {
		case <-started:
		case <-timeout:
			t.Errorf(`unable to start ssh-agent: %s`, stdErr)
			return
		}

		sock, err := net.Dial(`unix`, socket)
		So(err, ShouldBeNil)

		errA := sink.PassInputAgent(sock, protectedKeys, pass12345, stdErr)
		output := stdErr.Bytes()
		So(output, should.BeEmpty)
		So(errA, should.BeNil)

		keys, _ := connectOrPanic().List()
		So(3, should.Equal, len(keys))
		_ = keyring.Process.Kill()
		_ = os.Remove(socket)
	})
}
func startAndWait(started chan interface{}) {
	interval := 20*time.Microsecond
	i := 0 * time.Millisecond
	for {
		i+= interval
		if i > 1*time.Second {
			break
		}
		time.Sleep(interval)
		_, ne := os.Stat(socket)
		if ! os.IsNotExist(ne) {
			started <- 0
			break
		}
	}

}

func connectOrPanic() agent.ExtendedAgent {
	sock, err := net.Dial(`unix`, socket)
	if err != nil {
		panic(err)
	}
	return agent.NewClient(sock)
}
