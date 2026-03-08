//go:build windows

package process

import (
	"os/exec"
	"strconv"
	"syscall"
)

func SetupProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

func KillGroup(cmd *exec.Cmd) (err error) {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	pid := strconv.Itoa(cmd.Process.Pid)
	c := exec.Command("taskkill", "/T", "/F", "/PID", pid)
	err = c.Run()
	return
}

