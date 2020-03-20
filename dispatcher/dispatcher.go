package dispatcher

import (
	"context"
	"github.com/severgroup-tt/gopkg-app/client/sentry"
	logger "github.com/severgroup-tt/gopkg-logger"
	"sync"
)

// Dispatcher implementation of facade
type Dispatcher struct {
	processors map[Event][]EventProcessor
	bgCtx      context.Context
	// syncWG     sync.WaitGroup
	asyncWG sync.WaitGroup
	rwMutex sync.RWMutex
}

// AddListener add listener processors
func (d *Dispatcher) AddListener(listener EventListener) {
	d.rwMutex.Lock()
	defer d.rwMutex.Unlock()
	for name, processor := range listener.EventProcessors() {
		d.processors[name] = append(d.processors[name], processor...)
	}
}

// AddProcessor add processor for specific event
func (d *Dispatcher) AddProcessor(name Event, processor EventProcessor) {
	d.rwMutex.Lock()
	defer d.rwMutex.Unlock()
	d.processors[name] = append(d.processors[name], processor)
}

// Dispatch process event with all available processors in parallel and wait
func (d *Dispatcher) Dispatch(ctx context.Context, name Event, msg interface{}) {
	d.rwMutex.RLock()
	defer d.rwMutex.RUnlock()
	processorsCnt := len(d.processors[name])
	if processorsCnt == 0 {
		return
	}

	w := sync.WaitGroup{}
	w.Add(processorsCnt)
	for i := range d.processors[name] {
		processor := d.processors[name][i]

		go withRecover(ctx, func() {
			defer w.Done()
			if err := processor(ctx, msg); err != nil {
				logger.Error(ctx, "dispatcher Dispatch error on event: %s, error: %#v", name, err)
				sentry.Error(err)
			}
		})
	}
	w.Wait()
}

// Stop wait until all async processors done
func (d *Dispatcher) Stop() {
	d.rwMutex.Lock()
	defer d.rwMutex.Unlock()
	d.asyncWG.Wait()
}

// withRecover wrap function with panic catch
func withRecover(ctx context.Context, fn func()) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(ctx, "dispatcher-panic error: %v", err)
			sentry.Panic(err)
		}
	}()

	fn()
}
