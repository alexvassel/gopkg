package app

import (
	"github.com/utrack/clay/v2/transport/swagger"
	"google.golang.org/grpc"
	"net/http"
)

type PublicCloserFn func() error
type OptionFn func(a *App)

type PublicHandler struct {
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

func WithPublicHandler(method, pattern string, handlerFunc http.HandlerFunc) OptionFn {
	return func(a *App) {
		a.customPublicHandler = append(a.customPublicHandler, PublicHandler{Method: method, Pattern: pattern, HandlerFunc: handlerFunc})
	}
}

func WithPublicCloser(fn ...PublicCloserFn) OptionFn {
	return func(a *App) {
		a.customPublicCloser = append(a.customPublicCloser, fn...)
	}
}

func WithUnaryInterceptor(interceptor ...grpc.UnaryServerInterceptor) OptionFn {
	return func(a *App) {
		a.unaryInterceptor = append(a.unaryInterceptor, interceptor...)
	}
}

func WithPublicMiddleware(middleware ...func(http.Handler) http.Handler) OptionFn {
	return func(a *App) {
		a.publicMiddleware = append(a.publicMiddleware, middleware...)
	}
}

func WithPprof(enabled bool) OptionFn {
	return func(a *App) {
		a.customEnablePprof = enabled
	}
}

func WithSwaggerOption(o ...swagger.Option) OptionFn {
	return func(a *App) {
		a.customSwaggerOption = append(a.customSwaggerOption, o...)
	}
}
