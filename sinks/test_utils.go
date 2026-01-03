package sinks

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/joy-dx/relay/dto"
)

var stdoutMu sync.Mutex

func CaptureStdout(t *testing.T, fn func()) string {
	t.Helper()

	stdoutMu.Lock()
	defer stdoutMu.Unlock()

	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = orig

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()

	return buf.String()
}

type basicEvent struct {
	ref dto.EventRef
	msg string
}

func (e basicEvent) RelayChannel() dto.EventChannel { return "relay" }
func (e basicEvent) RelayType() dto.EventRef        { return e.ref }
func (e basicEvent) Message() string                { return e.msg }
func (e basicEvent) ToSlog() []slog.Attr            { return nil }
