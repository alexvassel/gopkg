package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/severgroup-tt/gopkg-app/middleware"
	"net"
	"net/http"
	"net/http/pprof"
	"syscall"
	"time"

	"github.com/severgroup-tt/gopkg-app/closer"
	"github.com/severgroup-tt/gopkg-app/metrics"
	swaggerui "github.com/severgroup-tt/gopkg-app/swagger"
	pkgtransport "github.com/severgroup-tt/gopkg-app/transport"
	pkgvalidator "github.com/severgroup-tt/gopkg-app/validator"
	validatorerr "github.com/severgroup-tt/gopkg-app/validator/errors"
	validatormw "github.com/severgroup-tt/gopkg-app/validator/middleware"

	"github.com/go-chi/chi"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	errors "github.com/severgroup-tt/gopkg-errors"
	errgrpc "github.com/severgroup-tt/gopkg-errors/grpc"
	errhttp "github.com/severgroup-tt/gopkg-errors/http"
	errmw "github.com/severgroup-tt/gopkg-errors/middleware"
	logger "github.com/severgroup-tt/gopkg-logger"
	"github.com/utrack/clay/v2/transport"
	"github.com/utrack/clay/v2/transport/swagger"
	"google.golang.org/grpc"
)

var (
	gracefulDelay   = time.Duration(3) * time.Second
	gracefulTimeout = time.Duration(10) * time.Second
)

type App struct {
	config Config

	httpServer        *chi.Mux
	httpListener      net.Listener
	httpAdminServer   *chi.Mux
	httpAdminListener net.Listener
	grpcServer        *grpc.Server
	grpcListener      net.Listener

	unaryInterceptor []grpc.UnaryServerInterceptor
	publicMiddleware []func(http.Handler) http.Handler

	publicCloser *closer.Closer

	customPublicHandler []PublicHandler
	customPublicCloser  []PublicCloserFn
	customSwaggerOption []swagger.Option
	customEnablePprof   bool
}

func NewApp(ctx context.Context, config Config, option ...OptionFn) (*App, error) {
	pkgtransport.Override(nil)
	metrics.AddBasicCollector(config.Name)

	a := &App{
		config:           config,
		unaryInterceptor: getDefaultUnaryInterceptor(config.Name),
		publicMiddleware: getDefaultPublicMiddleware(config.Version),
		publicCloser:     closer.New(syscall.SIGTERM, syscall.SIGINT),
	}
	if err := a.initServers(ctx); err != nil {
		return nil, err
	}

	for _, optFn := range option {
		optFn(a)
	}

	return a, nil
}

func (a *App) Run(impl ...transport.Service) {
	var descs []transport.ServiceDesc
	for _, i := range impl {
		descs = append(descs, i.GetDescription())
	}
	implDesc := transport.NewCompoundServiceDesc(descs...)
	implDesc.Apply(transport.WithUnaryInterceptor(grpc_middleware.ChainUnaryServer(
		append(a.unaryInterceptor)...,
	)))
	a.runServers(implDesc)
}

func GracefulDelay(serviceName string) {
	logger.Log(logger.App, serviceName+": waiting stop of traffic")
	time.Sleep(gracefulDelay)
	logger.Log(logger.App, serviceName+": shutting down")
}

func (a *App) runServers(impl *transport.CompoundServiceDesc) {
	if a.grpcServer != nil {
		a.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(a.unaryInterceptor...)))
		impl.RegisterGRPC(a.grpcServer)
		a.runGRPC()
	}

	if a.httpServer != nil {
		a.httpServer.Use(a.publicMiddleware...)
		impl.RegisterHTTP(a.httpServer)
		a.httpServer.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "text/html")
			text := a.config.Name + " " + a.config.Version
			_, _ = w.Write([]byte("<html><head><title>" + text + `</title></head><body><h1 style="text-align: center; margin-top: 200px; font-size: 50px;">` + text + "</h1></body></html>"))
		})
		a.httpServer.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "image/x-icon")
			data, _ := base64.StdEncoding.DecodeString("AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA//////////////////////////////////7+//3+/v/9/v7//v79//7++v/+/v3//f7+//3+/v8AAAAA//////////////////////////////////////7+/v/8/v3/+vfu//PPlv/vumz/8cF4//bmxv/9/v3//f79///////////////////////////////////////+/v7//P77//TGg//2s1X/+rRS//q0Uv/ws1v/+OvR//7+/f///////////////////////////////////////v7+//n05P/0slj/+rNT//mzU//6s1L/+rNT//PPmP/+/v7///////////////////////////////////////7+/v/48+P/87NX//i0Uv/5tFP/+LRS//mzVP/zz5X//v79///////////////////////////////////////+/v7/+/37//LDgf/4s1P/+LRT//m0Uv/3s1n/9erQ//3+/v///////////////////////////////////////P7+//3+/f/59+r/8syS//S4Z//zv3P/9ePF//z+/P/+/v3///////////////////////////////////////7+/v/+/v7//f79//z+/P/+/fn//f75//3+/f/+/v3//v7+//3+/v/8/f3/9vv7/7zq9/+d4Pf/uen3//T7/P/7/f3//f7+//v8/f/wz9j/6Ka0/+q0wP/68PP/+/7+//7+/v/7/v7/+f38/4rX9P9VyPX/Usj3/1HJ9f+F1fX/9fz9//3+/v/or7r/42B3/+dfdv/oX3f/3nSH//ns8P/9/v7//f3+/8zv+f9UyPX/Tsn3/1HI9/9QyPj/Usj1/8ft+P/58PP/32V6/+lfd//nX3f/5193/+dfd//nqrb//v7+//v9/f+66vj/Usj3/1HI9/9RyPf/Ucj3/1LI9/+x5vn/8uLm/+Bgd//nX3f/5193/+dfd//oX3f/5Zem//7+/v/9/v7/3vT6/1jK8/9Ryff/Ucn3/1LI9/9TyvT/2PL7//z4+v/gb4T/5l92/+dfd//nX3f/4193/+66xf/+/v7//v7+//z9/v+z5vj/XMvz/1XI9v9Zy/L/quL3//r9/v/9/v7/79LY/+Btgf/lX3b/4mF3/+OWpP/7+fv//v7+///////+/v7//P7+/+b4+//N8Pn/5fb7//v9/P/9/v7/+v7+//z8/f/68/b/79Xb//Ti5//8/f3//f7+//3+/v8AAAAA//////7+/v/7/v7//P7+//3+/v/9/v7/+v7+//z+/v/7/v7//P7+//3+///9/v7//P7+//z+/v8AAAAAgAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAEAAA==")
			_, _ = w.Write(data)
		})
		if a.customEnablePprof {
			logger.Log(logger.App, "PPROF enabled")
			a.addPprofHandler()
		}
		for _, h := range a.customPublicHandler {
			a.httpServer.Method(h.Method, h.Pattern, h.HandlerFunc)
		}
		a.runPublicHTTP()
	}

	if a.httpAdminServer != nil {
		a.initAdminHandlers(impl)
		a.runAdminHTTP()
	}

	for _, fn := range a.customPublicCloser {
		a.publicCloser.Add(fn)
	}

	// Wait signal and close all Scratch resources
	a.publicCloser.Wait()
	// Close all other resources from globalCloser
	closer.CloseAll()
}

func getDefaultUnaryInterceptor(appName string) []grpc.UnaryServerInterceptor {
	errConverters := []errors.ErrorConverter{
		validatorerr.Converter(),
		errhttp.Converter(appName),
		errgrpc.Converter(appName),
	}
	return []grpc.UnaryServerInterceptor{
		errmw.NewConvertErrorsServerInterceptor(errConverters, &metrics.CountError),
		validatormw.NewValidateServerInterceptor(pkgvalidator.New()),
		middleware.NewLogInterceptor(),
	}
}

func getDefaultPublicMiddleware(appVersion string) []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		middleware.NewCorsMiddleware(),
		middleware.NewRequestIdMiddleware(),
		middleware.NewLogMiddleware(),
		middleware.NewNoCacheMiddleware(),
		middleware.NewVersionMiddleware(appVersion),
	}
}

func (a *App) addPprofHandler() {
	a.httpServer.HandleFunc("/debug/pprof/", pprof.Index)
	a.httpServer.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	a.httpServer.HandleFunc("/debug/pprof/profile", pprof.Profile)
	a.httpServer.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	a.httpServer.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

func (a *App) initServers(ctx context.Context) error {
	logger.Log(logger.App, "App '%s' version '%s' in %s started",
		a.config.Name,
		a.config.Version,
		a.config.Env,
	)

	if a.config.Listener.HttpPort != 0 {
		logger.Log(logger.App, "Starting public HTTP listener at %s:%d", a.config.Listener.Host, a.config.Listener.HttpPort)
		httpListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.config.Listener.Host, a.config.Listener.HttpPort))
		if err != nil {
			return err
		}
		a.httpListener = httpListener
		a.httpServer = chi.NewMux()
	}

	if a.config.Listener.HttpAdminPort != 0 {
		logger.Log(logger.App, "Starting admin HTTP listener at %s:%d", a.config.Listener.Host, a.config.Listener.HttpAdminPort)
		httpAdminListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.config.Listener.Host, a.config.Listener.HttpAdminPort))
		if err != nil {
			return err
		}
		a.httpAdminListener = httpAdminListener
		a.httpAdminServer = chi.NewMux()
	}

	if a.config.Listener.GrpcPort != 0 {
		logger.Log(logger.App, "Starting GRPC listener at %s:%d", a.config.Listener.Host, a.config.Listener.GrpcPort)
		grpcListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.config.Listener.Host, a.config.Listener.GrpcPort))
		if err != nil {
			return err
		}
		a.grpcListener = grpcListener
	}

	return nil
}

func (a *App) initAdminHandlers(implDesc *transport.CompoundServiceDesc) {
	// metrics
	a.httpAdminServer.Mount("/metrics", metrics.Metrics())

	// swagger
	a.httpAdminServer.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", 301)
	})
	a.httpAdminServer.Mount("/docs/", http.StripPrefix("/docs", swaggerui.NewHTTPHandler()))
	a.httpAdminServer.Get("/docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger.json", 301)
	})
	a.httpAdminServer.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-MimeType", "application/json")
		o := []swagger.Option{
			swagger.WithHost(a.config.HostAdmin),
			swagger.WithTitle(a.config.Name),
			swagger.WithVersion(a.config.Version),
			pkgtransport.SetIntegerTypeForInt64(),
			pkgtransport.SetNameSnakeCase(),
		}
		o = append(o, a.customSwaggerOption...)
		_, _ = w.Write(implDesc.SwaggerDef(o...))
	})
}

func (a *App) runGRPC() {
	go func() {
		if err := a.grpcServer.Serve(a.grpcListener); err != nil {
			logger.Error(logger.App, "grpc: %s", err)
			a.publicCloser.CloseAll()
		}
	}()
	a.publicCloser.Add(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
		defer cancel()

		GracefulDelay("grpc")

		done := make(chan struct{})
		go func() {
			a.grpcServer.GracefulStop()
			close(done)
		}()
		select {
		case <-done:
			logger.Log(logger.App, "grpc: gracefully stopped")
		case <-ctx.Done():
			err := errors.Internal.Err(context.Background(), "grpc: error during shutdown server").
				WithLogKV("error", ctx.Err())
			a.grpcServer.Stop()
			return errors.Internal.Err(context.Background(), "grpc: force stopped").
				WithLogKV("error", err)
		}
		return nil
	})
}

func (a *App) runPublicHTTP() {
	publicServer := &http.Server{Handler: a.httpServer}
	go func() {
		if err := publicServer.Serve(a.httpListener); err != nil {
			logger.Log(logger.App, "http.public: %s", err)
			a.publicCloser.CloseAll()
		}
	}()
	a.publicCloser.Add(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
		defer cancel()

		GracefulDelay("http.public")

		publicServer.SetKeepAlivesEnabled(false)
		if err := publicServer.Shutdown(ctx); err != nil {
			return errors.Internal.Err(context.Background(), "http.public: error during shutdown").
				WithLogKV("error", err)
		}
		logger.Log(logger.App, "http.public: gracefully stopped")
		return nil
	})
}

func (a *App) runAdminHTTP() {
	adminServer := &http.Server{Handler: a.httpAdminServer}
	go func() {
		if err := adminServer.Serve(a.httpAdminListener); err != nil {
			logger.Log(logger.App, "admin.public: %s", err)
			a.publicCloser.CloseAll()
		}
	}()
	a.publicCloser.Add(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
		defer cancel()

		GracefulDelay("admin.public")

		adminServer.SetKeepAlivesEnabled(false)
		if err := adminServer.Shutdown(ctx); err != nil {
			return errors.Internal.Err(context.Background(), "admin.public: error during shutdown").
				WithLogKV("error", err)
		}
		logger.Log(logger.App, "admin.public: gracefully stopped")
		return nil
	})
}
