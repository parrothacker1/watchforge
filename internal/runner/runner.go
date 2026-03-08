package runner

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/parrothacker1/watchforge/internal/process"
)

type Runner struct {
	cmdStr string

	mu  sync.Mutex
	cmd *exec.Cmd
}

func New(cmd string) *Runner {
	return &Runner{
		cmdStr: cmd,
	}
}

func (r *Runner) Start() error {

	r.mu.Lock()
	defer r.mu.Unlock()

	cmd := exec.Command("sh", "-c", r.cmdStr)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	process.SetupProcessGroup(cmd)

	err := cmd.Start()
	if err != nil {
		return err
	}

	r.cmd = cmd

	go func() {
		cmd.Wait()
	}()

	fmt.Println("[watchforge] server started")

	return nil
}

func (r *Runner) Stop() error {

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cmd == nil || r.cmd.Process == nil {
		return nil
	}

	fmt.Println("[watchforge] stopping server")

	err := process.KillGroup(r.cmd)

	r.cmd = nil

	return err
}

func (r *Runner) Restart() error {

	err := r.Stop()
	if err != nil {
		return err
	}

	return r.Start()
}

