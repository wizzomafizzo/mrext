//go:build linux

package service

import (
	"fmt"
	"syscall"
)

func SetNice() error {
	if err := syscall.Setpriority(syscall.PRIO_PROCESS, 0, 1); err != nil {
		return fmt.Errorf("set process priority: %w", err)
	}
	return nil
}
