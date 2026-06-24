//go:build linux

package sflag

import (
	"os"
	"syscall"
	"unsafe"
)

func isTerminal(f *os.File) bool {
	var termios [256]byte
	// TCGETS on Linux
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		uintptr(0x5401),
		uintptr(unsafe.Pointer(&termios)),
	)
	return err == 0
}
