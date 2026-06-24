//go:build windows

package sflag

import (
	"os"
	"syscall"
)

func isTerminal(f *os.File) bool {
	var mode uint32
	// If GetConsoleMode succeeds, it is an interactive terminal window.
	// If it fails (returns an error), the output is being piped or redirected.
	err := syscall.GetConsoleMode(syscall.Handle(f.Fd()), &mode)
	return err == nil
}
