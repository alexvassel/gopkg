package app

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/severgroup-tt/gopkg-app/client/sentry"
	"github.com/severgroup-tt/gopkg-app/metrics"
	"github.com/severgroup-tt/gopkg-app/middleware"
	"github.com/severgroup-tt/gopkg-app/tracing"
	"github.com/utrack/clay/v2/transport/swagger"
	"google.golang.org/grpc"
	"net/http"
	"time"
)

type PublicCloserFn func() error
type OptionFn func(a *App) error

type PublicHandler struct {
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

func WithPublicHandler(method, pattern string, handlerFunc http.HandlerFunc) OptionFn {
	return func(a *App) error {
		a.customPublicHandler = append(a.customPublicHandler, PublicHandler{Method: method, Pattern: pattern, HandlerFunc: handlerFunc})
		return nil
	}
}

func WithPublicCloser(fn ...PublicCloserFn) OptionFn {
	return func(a *App) error {
		a.customPublicCloser = append(a.customPublicCloser, fn...)
		return nil
	}
}

func WithUnaryPrependInterceptor(interceptor ...grpc.UnaryServerInterceptor) OptionFn {
	return func(a *App) error {
		a.unaryInterceptor = append(interceptor, a.unaryInterceptor...)
		return nil
	}
}

func WithUnaryAppendInterceptor(interceptor ...grpc.UnaryServerInterceptor) OptionFn {
	return func(a *App) error {
		a.unaryInterceptor = append(a.unaryInterceptor, interceptor...)
		return nil
	}
}

func WithPublicMiddleware(middleware ...func(http.Handler) http.Handler) OptionFn {
	return func(a *App) error {
		a.publicMiddleware = append(a.publicMiddleware, middleware...)
		return nil
	}
}

func WithPprof(enabled bool) OptionFn {
	return func(a *App) error {
		a.customEnablePprof = enabled
		return nil
	}
}

func WithSwaggerOption(o ...swagger.Option) OptionFn {
	return func(a *App) error {
		a.customSwaggerOption = append(a.customSwaggerOption, o...)
		return nil
	}
}

func WithMetrics(metric ...prometheus.Collector) OptionFn {
	return func(a *App) error {
		metrics.AddCollector(metric...)
		return nil
	}
}

func WithTracer(addr string) OptionFn {
	return func(a *App) error {
		tracerCloser, err := tracing.InitTracer(a.config.Name, addr)
		if err != nil {
			return err
		}
		a.publicCloser.Add(func() error {
			return tracerCloser.Close()
		})
		a.unaryInterceptor = append(a.unaryInterceptor,
			middleware.NewUnaryTracingInterceptor(tracing.GetTracer()),
		)
		return nil
	}
}

func WithSentry(project, token string, timeout time.Duration) OptionFn {
	return func(a *App) error {
		if err := sentry.Init(project, token, timeout); err != nil {
			return err
		}
		return nil
	}
}

func WithFavicon(favicon []byte) OptionFn {
	return func(a *App) error {
		a.favicon = favicon
		return nil
	}
}

func WithAdminURLPrefix(prefix string) OptionFn {
	return func(a *App) error {
		a.adminURLPrefix = prefix
		return nil
	}
}
