package builder

import (
	"context"
	"os"
	"os/exec"
	"sync"

	"github.com/parrothacker1/watchforge/internal/config"
	"github.com/parrothacker1/watchforge/internal/logger"
)

type Builder struct {
	mu     sync.Mutex
	cancel context.CancelFunc
	ctx    context.Context
}

func New(ctx context.Context) *Builder {
	return &Builder{
		ctx: ctx,
	}
}

func (b *Builder) Build() error {
	b.mu.Lock()
	if b.cancel != nil {
		b.cancel()
	}
	ctx, cancel := context.WithCancel(b.ctx)
	b.cancel = cancel
	b.mu.Unlock()
	logger.Log.Info("running builder")
	buildCmd := config.GetConfig().Build.Command
	cmd := exec.CommandContext(ctx, "sh", "-c", buildCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (b *Builder) Cancel() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.cancel != nil {
		logger.Log.Info("stopping builder")
		b.cancel()
	}
}
