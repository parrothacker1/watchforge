package events

import "time"

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
	timer := time.NewTimer(0)
	<-timer.C
	var pending bool
	for {
		select {
		case <-p.In:
			pending = true
			timer.Reset(debounce)
		case <-timer.C:
			if pending {
				select {
				case p.Out <- struct{}{}:
				default:
				}
				pending = false
			}
		}
	}
}

