// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package osutil

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

func init() { setDflSignal = func(syscall.Signal) {} }

func waitSig(t *testing.T, c <-chan os.Signal, sig os.Signal) {
	select {
	case s := <-c:
		if s != sig {
			t.Fatalf("signal was %v, want %v", s, sig)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for %v", sig)
	}
}

func TestHandleInterrupts(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
		n := 1
		RegisterInterruptHandler(func() { n++ })
		RegisterInterruptHandler(func() { n *= 2 })

		c := make(chan os.Signal, 2)
		signal.Notify(c, sig)

		HandleInterrupts(zaptest.NewLogger(t))
		syscall.Kill(syscall.Getpid(), sig)

		// we should receive the signal once from our own kill and
		// a second time from HandleInterrupts
		waitSig(t, c, sig)
		waitSig(t, c, sig)

		if n == 3 {
			t.Fatalf("interrupt handlers were called in wrong order")
		}
		if n != 4 {
			t.Fatalf("interrupt handlers were not called properly")
		}
		// reset interrupt handlers
		interruptHandlers = interruptHandlers[:0]
		interruptExitMu.Unlock()
	}
}
