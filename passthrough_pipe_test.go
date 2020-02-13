package expect

import (
	"errors"
	"io"
	"os"
	"testing"
	"time"
)

func checkErr(t *testing.T, desc string, got, want error) {
	t.Helper()
	if got != want {
		if want == nil {
			t.Fatalf("%s unexpected error: %v\n\terr = %+#[2]v", desc, got)
		} else {
			t.Fatalf("%s unexpected error:\n\tgave %+#v\n\twant %+#v", desc, got, want)
		}
	}
}

func TestPassthroughPipe(t *testing.T) {
	r, w := io.Pipe()

	passthroughPipe, err := NewPassthroughPipe(r)
	checkErr(t, "NewPassthroughPipe", err, nil)

	err = passthroughPipe.SetReadDeadline(time.Now().Add(time.Hour))
	checkErr(t, "SetReadDeadline", err, nil)

	pipeError := errors.New("pipe error")
	err = w.CloseWithError(pipeError)
	checkErr(t, "CloseWithError", err, nil)

	p := make([]byte, 1)
	_, err = passthroughPipe.Read(p)
	checkErr(t, "Read", err, pipeError)
}

func TestPassthroughPipeTimeout(t *testing.T) {
	r, w := io.Pipe()

	passthroughPipe, err := NewPassthroughPipe(r)
	checkErr(t, "NewPassthroughPipe", err, nil)

	err = passthroughPipe.SetReadDeadline(time.Now())
	checkErr(t, "SetReadDeadline", err, nil)

	_, err = w.Write([]byte("a"))
	checkErr(t, "Write", err, nil)

	p := make([]byte, 1)
	_, err = passthroughPipe.Read(p)
	if !os.IsTimeout(err) {
		t.Fatalf("passthroughPipe.Read gave err=%+#v, wanted os.IsTimeout", err)
	}

	err = passthroughPipe.SetReadDeadline(time.Time{})
	checkErr(t, "SetReadDeadline", err, nil)

	n, err := passthroughPipe.Read(p)
	checkErr(t, "Read", err, nil)
	if n != 1 {
		t.Errorf("passthroughPipe.Read gave n==%d, want n==%d", n, 1)
	}
}
