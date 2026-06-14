//go:build !darwin && !linux

package sflag

import "io"

func isWriterTerminal(_ io.Writer) bool { return false }
