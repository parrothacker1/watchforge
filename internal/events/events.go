package events

import (
	"path/filepath"
	"time"

	"github.com/parrothacker1/watchforge/internal/logger"
)

type Event struct {
	Path string
}

type Batch struct {
	Paths []string
}

type Processor struct {
	In  chan Event
	Out chan Batch
}

func NewProcessor(buffer int) *Processor {
	return &Processor{
		In:  make(chan Event, buffer),
		Out: make(chan Batch, 1),
	}
}

func (p *Processor) Run(debounce time.Duration) {
	timer := time.NewTimer(debounce)
	timer.Stop()
	var pending bool
	paths := make(map[string]struct{})
	for {
		select {
		case ev := <-p.In:
			pending = true
			ev.Path = filepath.Clean(ev.Path)
			paths[ev.Path] = struct{}{}
			logger.Log.Debug("event queued", "path", ev.Path)
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(debounce)
		case <-timer.C:
			if pending {
				list := make([]string, 0, len(paths))
				for p := range paths {
					list = append(list, p)
				}
				logger.Log.Debug(
					"event batch flushed",
					"count", len(list),
				)
				select {
				case p.Out <- Batch{Paths: list}:
				default:
				}
				paths = make(map[string]struct{})
				pending = false
			}
		}
	}
}
