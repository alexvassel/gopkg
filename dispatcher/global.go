package dispatcher

import "context"

var instance *Dispatcher

func init() {
	instance = &Dispatcher{
		processors: make(map[Event][]EventProcessor),
		bgCtx:      context.Background(),
	}
}

// Get returns global dispatcher
func Get() EventDispatcher {
	return instance
}

// SetBackgroundContext ...
func SetBackgroundContext(ctx context.Context) {
	instance.bgCtx = ctx
}

// Dispatch dispatch message using global dispatcher
func Dispatch(ctx context.Context, name Event, msg interface{}) {
	instance.Dispatch(ctx, name, msg)
}
