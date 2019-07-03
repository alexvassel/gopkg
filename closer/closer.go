package closer

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
)

// GlobalCloser automatically calls when app is terminating
var globalCloser = New()

// Add adds `func() error` callback to the globalCloser
func Add(f ...func() error) {
	globalCloser.Add(f...)
}

func Wait() {
	globalCloser.Wait()
}

func CloseAll() {
	globalCloser.CloseAll()
}

type Closer struct {
	sync.Mutex
	once  sync.Once
	done  chan struct{}
	funcs []func() error
}

// New returns new Closer, if []os.Signal is specified Closer will automatically call CloseAll when one of signals is received from OS
func New(sig ...os.Signal) *Closer {
	c := &Closer{done: make(chan struct{})}
	if len(sig) > 0 {
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, sig...)
			<-ch
			signal.Stop(ch)
			c.CloseAll()
		}()
	}
	return c
}

func (c *Closer) Add(f ...func() error) {
	c.Lock()
	c.funcs = append(c.funcs, f...)
	c.Unlock()
}

func (c *Closer) Wait() {
	select {
	case <-c.done:
	}
}

func (c *Closer) CloseAll() {
	c.once.Do(func() {
		defer close(c.done)

		c.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.Unlock()

		// call all Closer funcs async
		errs := make(chan error, len(funcs))
		for _, f := range funcs {
			go func(f func() error) {
				errs <- f()
			}(f)
		}

		for i := 0; i < cap(errs); i++ {
			err := <-errs
			if err := err; err != nil {
				fmt.Println(err)
			}
		}
	})
}
