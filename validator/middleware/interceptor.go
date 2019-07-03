package middleware

import (
	"context"

	"google.golang.org/grpc"
	"gopkg.in/go-playground/validator.v9"
)

// NewValidateServerInterceptor wrap endpoint with validation middleware
func NewValidateServerInterceptor(validate *validator.Validate) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		err = validate.Struct(req)
		if err != nil {
			return
		}

		return handler(ctx, req)
	}
}
