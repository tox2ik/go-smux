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

const socket = `testdata/keyring.sock`

var protectedKeys = []string{
	`testdata/jwf.ossh`,
	`testdata/jwf.pem`,
	`testdata/jwf.ec19`,
}

func TestSshAgent_NotRunnning(t *testing.T) {
	t.Skipf("not implemented.")
}

func TestSshAgent_ImportOpenKey(t *testing.T) {
	t.Skipf("not implemented.")
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

func TestCat_PipeLines(t *testing.T) {
	Convey("Pipe input to any binary", t, func() {
		Convey("Feed a single line to cat -n", func(){
			buf := bytes.NewBuffer(make([]byte, 0))
			cmd := execCommand("cat", []string{"-n"})
			cmd.Stdout = buf
			_ = sink.PassInputString(cmd, "Gewürz Gurken\n")
			So(string(buf.Bytes()), ShouldEqual, "     1\tGewürz Gurken\n")
		})

		Convey("Feed two lines to cat -n", func(){
			buf := bytes.NewBuffer(make([]byte, 0))
			cmd := execCommand("cat", []string{"-n"})
			cmd.Stdout = buf
			_ = sink.PassInputString(cmd, "one\ntwo\n")
			So(string(buf.Bytes()), ShouldEqual, "     1\tone\n     2\ttwo\n")
		})
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


/***

no agent running.


$ ssh-add-keys
passphrase:
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x20 pc=0x5922f1]

goroutine 1 [running]:
golang.org/x/crypto/ssh/agent.(*client).callRaw(0xc00018a260, 0xc0001bca80, 0x759, 0xa42, 0x0, 0x0, 0x0, 0x0, 0x0)
	/home/jaroslav/src/go/pkg/mod/golang.org/x/crypto@v0.0.0-20200604202706-70a84ac30bf9/ssh/agent/client.go:342 +0x131
golang.org/x/crypto/ssh/agent.(*client).call(0xc00018a260, 0xc0001bca80, 0x759, 0xa42, 0xc0001be070, 0xc0001bca80, 0x759, 0xa42)
	/home/jaroslav/src/go/pkg/mod/golang.org/x/crypto@v0.0.0-20200604202706-70a84ac30bf9/ssh/agent/client.go:321 +0x50
golang.org/x/crypto/ssh/agent.(*client).insertKey(0xc00018a260, 0x5d1560, 0xc0001a03c0, 0x7fffde0f505b, 0x2a, 0x0, 0x0, 0x0, 0x791e01, 0xc00018c5d0)
	/home/jaroslav/src/go/pkg/mod/golang.org/x/crypto@v0.0.0-20200604202706-70a84ac30bf9/ssh/agent/client.go:598 +0x286
golang.org/x/crypto/ssh/agent.(*client).Add(0xc00018a260, 0x5d1560, 0xc0001a03c0, 0x0, 0x7fffde0f505b, 0x2a, 0x0, 0x0, 0x0, 0x0, ...)
	/home/jaroslav/src/go/pkg/mod/golang.org/x/crypto@v0.0.0-20200604202706-70a84ac30bf9/ssh/agent/client.go:659 +0x14b
genja.org/smux/sink.PassInputAgent(0x0, 0x0, 0xc0001a02b0, 0x5, 0x5, 0x630ae0, 0xc0000388d0, 0x7f6ba0908088, 0xc00000eba0, 0xc000065f08, ...)
	/home/jaroslav/workspace/smux/sink/ssh_agent.go:37 +0x5b4
main.main()
	/home/jaroslav/workspace/smux/smux.go:53 +0x54a
Error connecting to agent: No such file or directory

**/
