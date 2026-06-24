//go:build darwin

package sflag

import (
	"os"
	"syscall"
	"unsafe"
)

func isTerminal(f *os.File) bool {
	var termios [256]byte
	// TIOCGETA on macOS
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		uintptr(0x404c7413),
		uintptr(unsafe.Pointer(&termios)),
	)
	return err == 0
}
