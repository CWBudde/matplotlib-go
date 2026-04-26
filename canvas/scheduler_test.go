package canvas

import (
	"testing"
	"time"
)

func TestSynchronousEventLoopCallSoon(t *testing.T) {
	called := false
	if err := (SynchronousEventLoop{}).CallSoon(func() error {
		called = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("callback was not called")
	}
}

func TestTimerLifecycle(t *testing.T) {
	timer := NewTimer(time.Hour, func() error { return nil })
	if timer.Running() {
		t.Fatal("timer running before Start")
	}
	if err := timer.Start(); err != nil {
		t.Fatal(err)
	}
	if !timer.Running() {
		t.Fatal("timer not running after Start")
	}
	if err := timer.Stop(); err != nil {
		t.Fatal(err)
	}
	if timer.Running() {
		t.Fatal("timer running after Stop")
	}
}
