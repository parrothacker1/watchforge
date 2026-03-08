package builder

import (
	"context"
	"os"
	"os/exec"
	"sync"
)

type Builder struct {
	buildCmd string

	mu     sync.Mutex
	cancel context.CancelFunc
	ctx    context.Context
}

func New(cmd string, ctx context.Context) *Builder {
	return &Builder{
		buildCmd: cmd,
		ctx:      ctx,
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
	cmd := exec.CommandContext(ctx, "sh", "-c", b.buildCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (b *Builder) Cancel() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.cancel != nil {
		b.cancel()
	}
}
