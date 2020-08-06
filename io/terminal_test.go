
package io

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTerminal_ReadWritten(t *testing.T) {

	Convey("Read & write to same buffer", t, func() {
		mw := NewMultiWriter()
		_,_ = mw.WriteString("Hello")
		err := mw.Flush()
		println(err)
		So(string(mw.Bytes()), ShouldEqual, "Hello")
		So(mw.String(), ShouldEqual, "Hello")
	})

}

