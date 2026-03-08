package runner

import (
	"context"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/parrothacker1/watchforge/internal/config"
	"github.com/parrothacker1/watchforge/internal/logger"
	"github.com/parrothacker1/watchforge/internal/process"
)

type Runner struct {
	ctx context.Context

	mu        sync.Mutex
	cmd       *exec.Cmd
	lastStart time.Time
	crashes   int
}

func New(ctx context.Context) *Runner {
	return &Runner{
		ctx: ctx,
	}
}

func (r *Runner) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	logger.Log.Info("starting runner")
	cmdStr := config.GetConfig().Run.Command
	cmd := exec.CommandContext(r.ctx, "sh", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	process.SetupProcessGroup(cmd)
	if err := cmd.Start(); err != nil {
		logger.Log.Error("failed to start runner", "error", err)
		return err
	}
	r.lastStart = time.Now()
	r.cmd = cmd

	go func() {
		cfg := config.GetConfig()
		err := cmd.Wait()
		r.mu.Lock()
		defer r.mu.Unlock()

		runTime := time.Since(r.lastStart)

		if runTime < time.Duration(cfg.Runner.CrashWindow) {
			r.crashes++
			logger.Log.Warn("runner crashed quickly", "runtime", runTime, "crashes", r.crashes)
			if r.crashes >= cfg.Runner.MaxCrashes {
				logger.Log.Warn("crash loop detected, delaying restart", "delay", cfg.Runner.RestartDelay)
				time.Sleep(time.Duration(cfg.Runner.RestartDelay))
			}
		} else {
			r.crashes = 0
		}

		if err != nil {
			logger.Log.Debug("runner exited", "error", err)
		} else {
			logger.Log.Debug("runner exited")
		}
	}()
	logger.Log.Info("runner started", "pid", cmd.Process.Pid)
	return nil
}

func (r *Runner) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd == nil || r.cmd.Process == nil {
		return nil
	}
	logger.Log.Info("stopping runner")
	err := process.KillGroup(r.cmd)
	if err != nil {
		logger.Log.Error("failed to stop runner", "error", err)
		return err
	}
	r.cmd = nil
	logger.Log.Info("runner stopped")
	return nil
}

func (r *Runner) Restart() error {
	logger.Log.Info("restarting runner")
	if err := r.Stop(); err != nil {
		return err
	}
	return r.Start()
}
