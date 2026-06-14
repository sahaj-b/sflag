//go:build !darwin && !linux

package sflag

func isTerminal() bool { return false }
