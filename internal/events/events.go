package events

import (
	"time"

	"github.com/parrothacker1/watchforge/internal/logger"
)

type Event struct {
	Path string
}

type Processor struct {
	In  chan Event
	Out chan struct{}
}

func NewProcessor(buffer int) *Processor {
	return &Processor{
		In:  make(chan Event, buffer),
		Out: make(chan struct{}, 1),
	}
}

func (p *Processor) Run(debounce time.Duration) {
	timer := time.NewTimer(debounce)
	timer.Stop()
	var pending bool
	var lastPath string
	count := 0
	for {
		select {
		case ev := <-p.In:
			lastPath = ev.Path
			pending = true
			count++
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
				logger.Log.Debug(
					"event batch flushed",
					"last_path", lastPath,
					"count", count,
				)
				select {
				case p.Out <- struct{}{}:
				default:
				}
				pending = false
				count = 0
				lastPath = ""
			}
		}
	}
}
