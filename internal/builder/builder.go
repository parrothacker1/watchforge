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
}

func New(cmd string) *Builder {
	return &Builder{
		buildCmd: cmd,
	}
}

func (b *Builder) Build() error {
	b.mu.Lock()
	if b.cancel != nil {
		b.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
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

