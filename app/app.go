package app

import (
	"context"
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
