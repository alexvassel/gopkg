package grpc

import (
	"context"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	logger "github.com/severgroup-tt/gopkg-logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	AppNameHeader    = "x-app-name"
	AppVersionHeader = "x-app-version"
)

// DefaultUnaryInterceptors ...
func defaultUnaryInterceptors(opts Option) []grpc.UnaryClientInterceptor {
	return []grpc.UnaryClientInterceptor{
		NewAppInfoUnaryInterceptor(opts.AppName, opts.AppVersion),
		NewRetryUnaryInterceptor(opts.Service, opts.MaxRetry, opts.RetryDelay),
	}
}

// DefaultStreamInterceptors ...
func defaultStreamInterceptors(opts Option) []grpc.StreamClientInterceptor {
	return []grpc.StreamClientInterceptor{
		NewAppInfoStreamInterceptor(opts.AppName, opts.AppVersion),
	}
}

func NewRetryUnaryInterceptor(service string, attempts uint, delay time.Duration) grpc.UnaryClientInterceptor {
	if delay < time.Millisecond {
		delay = DefaultRetryDelay
	}
	if attempts < 1 {
		attempts = DefaultMaxRetry
	}
	retryInterceptor := grpc_retry.UnaryClientInterceptor(grpc_retry.WithMax(attempts))

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		retryOpts := append(opts, grpc_retry.WithBackoff(func(attempt uint) time.Duration {
			logger.Error(ctx, "Failed to fetch GRPC data. "+
				"Service: %s, Host: %s, Url: %s, Request: %v, RetryAttempt: %d ",
				service, cc.Target(), method, req, attempt)

			return delay
		}))
		return retryInterceptor(ctx, method, req, reply, cc, invoker, retryOpts...)
	}
}

func NewAppInfoUnaryInterceptor(sourceAppName string, sourceAppVersion string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, AppNameHeader, sourceAppName)
		ctx = metadata.AppendToOutgoingContext(ctx, AppVersionHeader, sourceAppVersion)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func NewAppInfoStreamInterceptor(sourceAppName string, sourceAppVersion string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx = metadata.AppendToOutgoingContext(ctx, AppNameHeader, sourceAppName)
		ctx = metadata.AppendToOutgoingContext(ctx, AppVersionHeader, sourceAppVersion)
		return streamer(ctx, desc, cc, method, opts...)
	}
}
