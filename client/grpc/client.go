package grpc

import (
	grpcmw "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func NewClient(opt Option, extra ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(
		opt.Addr,
		append([]grpc.DialOption{
			grpc.WithInsecure(),
			grpc.WithUserAgent(opt.AppName),
			grpc.WithUnaryInterceptor(grpcmw.ChainUnaryClient(defaultUnaryInterceptors(opt)...)),
			grpc.WithStreamInterceptor(grpcmw.ChainStreamClient(defaultStreamInterceptors(opt)...)),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				PermitWithoutStream: true, // most likely it does not work without proper setup on server side
			}),
		}, extra...)...)
}
