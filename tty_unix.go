//go:build darwin || linux

package sflag

import (
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

const (
	tioWinszLinux  = 0x5413
	tioWinszDarwin = 0x40087468
)

type winsize struct{ row, col, xpixel, ypixel uint16 }

func isTerminal() bool {
	var ws winsize
	req := uintptr(tioWinszLinux)
	if runtime.GOOS == "darwin" {
		req = uintptr(tioWinszDarwin)
	}
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, os.Stderr.Fd(), req, uintptr(unsafe.Pointer(&ws)))
	return err == 0
}
