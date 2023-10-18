// Package publisher manages the publishing of metrics.
package publisher

import (
	"log/slog"
	"sync"
	"time"
)

// =============================================================================

// Collector defines a contract a collector must support
// so a consumer can retrieve metrics.
type Collector interface {
	Collect() (map[string]any, error)
}

// =============================================================================

// Publisher defines a handler function that will be called
// on each interval.
type Publisher func(map[string]any) error

// Publish provides the ability to receive metrics
// on an interval.
type Publish struct {
	log       *slog.Logger
	collector Collector
	publisher []Publisher
	wg        sync.WaitGroup
	timer     *time.Timer
	shutdown  chan struct{}
}

// New creates a Publish for consuming and publishing metrics.
func New(log *slog.Logger, collector Collector, interval time.Duration, publisher ...Publisher) (*Publish, error) {
	p := Publish{
		log:       log,
		collector: collector,
		publisher: publisher,
		timer:     time.NewTimer(interval),
		shutdown:  make(chan struct{}),
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			p.timer.Reset(interval)
			select {
			case <-p.timer.C:
				if err := p.update(); err != nil {
					log.Error("update", "status", "update data", "msg", err)
				}
			case <-p.shutdown:
				return
			}
		}
	}()

	return &p, nil
}

// Stop is used to shut down the goroutine collecting metrics.
func (p *Publish) Stop() {
	close(p.shutdown)
	p.wg.Wait()
}

// update pulls the metrics and publishes them to the specified system.
func (p *Publish) update() error {
	data, err := p.collector.Collect()
	if err != nil {
		p.log.Error("publish", "status", "collect data", "msg", err)
		return err
	}

	for _, pub := range p.publisher {
		err = pub(data)
		if err != nil {
			p.log.Error("publish", "status", "collect data", "publisher", pub, "msg", err)
		}
	}

	return nil
}
