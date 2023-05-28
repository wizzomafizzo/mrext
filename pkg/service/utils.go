package service

import "syscall"

func SetNice() error {
	return syscall.Setpriority(syscall.PRIO_PROCESS, 0, 1)
}
