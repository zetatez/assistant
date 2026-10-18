// Package dwmblocknotify surfaces short-lived status messages. On Arch Linux
// the message is pushed to the dwmblocks notifier daemon; on macOS it is shown
// as a system notification. The package name is kept for compatibility with
// the existing dwm-oriented call sites.
package dwmblocknotify

import (
	"fmt"
	"os"
	"time"

	"assistant/internal/platform"
)

// POST asynchronously enqueues a message (shown after previous messages expire).
func POST(msg string, ttl time.Duration) {
	go func() {
		if err := platform.Get().NotifyQueue(msg, ttl); err != nil {
			logErr("POST", err)
		}
	}()
}

// PUT asynchronously preempts the current message.
func PUT(msg string, ttl time.Duration) {
	go func() {
		if err := platform.Get().Notify(msg, ttl); err != nil {
			logErr("PUT", err)
		}
	}()
}

// DELETE asynchronously clears the current message and the queue.
func DELETE() {
	go func() {
		if err := platform.Get().NotifyClear(); err != nil {
			logErr("DELETE", err)
		}
	}()
}

func logErr(method string, err error) {
	fmt.Fprintf(os.Stderr, "dwmblocknotify %s failed: %v\n", method, err)
}
