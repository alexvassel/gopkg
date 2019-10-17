package grpc

import (
	grpcmw "github.com/grpc-ecosystem/go-grpc-middleware"
	mw "github.com/severgroup-tt/gopkg-app/client/grpc/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// NewClient ...
func NewClient(opt Option, extra ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(
		opt.Addr,
		append(extra,
			grpc.WithInsecure(),
			grpc.WithUserAgent(opt.AppName),
			grpc.WithUnaryInterceptor(grpcmw.ChainUnaryClient(defaultUnaryInterceptors(opt)...)),
			grpc.WithStreamInterceptor(grpcmw.ChainStreamClient(defaultStreamInterceptors(opt)...)),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				PermitWithoutStream: true, // most likely it does not work without proper setup on server side
			}),
		)...)
}

// DefaultUnaryInterceptors ...
func defaultUnaryInterceptors(opts Option) []grpc.UnaryClientInterceptor {
	return []grpc.UnaryClientInterceptor{
		mw.NewAppInfoUnaryInterceptor(opts.AppName, opts.AppVersion),
		mw.NewLogUnaryInterceptor(opts.Service),
		mw.NewRetryUnaryInterceptor(opts.Service, opts.MaxRetry, opts.RetryDelay),
	}
}

// DefaultStreamInterceptors ...
func defaultStreamInterceptors(opts Option) []grpc.StreamClientInterceptor {
	return []grpc.StreamClientInterceptor{
		mw.NewAppInfoStreamInterceptor(opts.AppName, opts.AppVersion),
		mw.NewLogStreamInterceptor(opts.Service),
	}
}
