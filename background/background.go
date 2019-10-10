package background

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	database "github.com/severgroup-tt/gopkg-database"
	errors "github.com/severgroup-tt/gopkg-errors"
	logger "github.com/severgroup-tt/gopkg-logger"
)

type job struct {
	dbConn    database.IClient
	Name      string
	processor Processor
	cancel    context.CancelFunc
	stopChan  chan bool
	TickRate  time.Duration
	Timeout   time.Duration
	options   []OptionFn
}

func NewJob(name string, tickRate, timeout time.Duration, processor Processor, opts ...OptionFn) IService {
	return &job{
		Name:      name,
		processor: processor,
		TickRate:  tickRate,
		Timeout:   timeout,
		options:   opts,
	}
}

func (j *job) Run(ctx context.Context) error {
	defer close(j.stopChan)
	for _, o := range j.options {
		o(j)
	}
	ctx, j.cancel = context.WithCancel(ctx)
	logger.Log(ctx, "Startup background %s, tick-rate %s, timeout %s", j.Name, j.TickRate, j.Timeout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-time.After(j.TickRate):
			logger.Log(ctx, "Start background %s", j.Name)
			if j.dbConn != nil {
				ctx = database.NewContext(ctx, j.dbConn)
			}
			withTimeout(ctx, j.Timeout, func(fnCtx context.Context) {
				err := tryProcess(fnCtx, j.processor)
				if j.isErrContext(err) {
					err = nil
				}
				if err != nil {
					logger.Error(fnCtx, "process failed, background %s, error %s", j.Name, err)
				}
			})
		}
	}
}

func withTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context)) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	doneChan := make(chan bool, 1)

	go func() {
		fn(ctx)
		close(doneChan)
	}()

	select {
	case <-doneChan:
	case <-ctx.Done():
	}

	cancel()
}

func tryProcess(ctx context.Context, processor Processor) (err error) {
	defer func() {
		cause := recover()
		if cause == nil {
			return
		}

		stack := make([]string, 0, 20)
		rawStack := make([]uintptr, 20)
		stackLength := runtime.Callers(4, rawStack)
		for _, pc := range rawStack[:stackLength] {
			fn := runtime.FuncForPC(pc)
			if fn == nil {
				continue
			}

			name := fn.Name()
			_, lineNumber := fn.FileLine(pc)
			stack = append(stack, fmt.Sprintf("%s:%v", name, lineNumber))
		}
		stackStr := strings.Join(stack, "; ")

		switch panicErr := cause.(type) {
		case error:
			err = errors.Internal.ErrWrap(ctx, "application panic", panicErr).WithLogKV("stack", stackStr)
		case string:
			err = errors.Internal.Err(ctx, panicErr).WithLogKV("stack", stackStr)
		default:
			err = errors.Internal.Err(ctx, fmt.Sprint(panicErr)).WithLogKV("stack", stackStr)
		}
	}()

	err = processor(ctx)
	return
}

func (j *job) isErrContext(err error) bool {
	return err == context.Canceled || err == context.DeadlineExceeded
}

func (j *job) Stop() {
	if j.cancel != nil {
		j.cancel()
		j.cancel = nil
	}
	<-j.stopChan
}

type IService interface {
	Run(ctx context.Context) error
	Stop()
}

type Processor func(context.Context) error
