package runner

import (
	"context"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/parrothacker1/watchforge/internal/logger"
	"github.com/parrothacker1/watchforge/internal/process"
)

type Runner struct {
	cmdStr string
	ctx    context.Context

	mu  sync.Mutex
	cmd *exec.Cmd

	lastStart time.Time
	crashes   int
}

func New(cmd string, ctx context.Context) *Runner {
	return &Runner{
		ctx:    ctx,
		cmdStr: cmd,
	}
}

func (r *Runner) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	logger.Log.Info("starting server")
	cmd := exec.CommandContext(r.ctx, "sh", "-c", r.cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	process.SetupProcessGroup(cmd)
	if err := cmd.Start(); err != nil {
		logger.Log.Error("failed to start server", "error", err)
		return err
	}
	r.lastStart = time.Now()
	r.cmd = cmd

	go func() {
		err := cmd.Wait()
		r.mu.Lock()
		defer r.mu.Unlock()
		runTime := time.Since(r.lastStart)
		if runTime < 2*time.Second {
			r.crashes++
			logger.Log.Warn(
				"server crashed quickly",
				"runtime", runTime,
				"crashes", r.crashes,
			)
			if r.crashes >= 3 {
				logger.Log.Warn(
					"crash loop detected, delaying restart",
					"delay", "2s",
				)
				time.Sleep(2 * time.Second)
			}
		} else {
			r.crashes = 0
		}
		if err != nil {
			logger.Log.Debug("server exited", "error", err)
		} else {
			logger.Log.Debug("server exited")
		}
	}()
	logger.Log.Info("server started", "pid", cmd.Process.Pid)
	return nil
}

func (r *Runner) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd == nil || r.cmd.Process == nil {
		return nil
	}
	logger.Log.Info("stopping server")
	err := process.KillGroup(r.cmd)
	if err != nil {
		logger.Log.Error("failed to stop server", "error", err)
		return err
	}
	r.cmd = nil
	logger.Log.Info("server stopped")
	return nil
}

func (r *Runner) Restart() error {
	logger.Log.Info("restarting server")
	if err := r.Stop(); err != nil {
		return err
	}
	return r.Start()
}
