package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
)

const (
	requestIDContextName = "request-id"
	requestIdHeaderName  = "X-Request-Id"
)

func SetRequestUuid(ctx context.Context) context.Context {
	b := make([]byte, 16)
	_, _ = rand.Read(b)

	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return context.WithValue(ctx, requestIDContextName, uuid)
}

func GetRequestId(ctx context.Context) string {
	uuid, _ := ctx.Value(requestIDContextName).(string)
	return uuid
}

func NewRequestIdMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(SetRequestUuid(r.Context()))
			w.Header().Add(requestIdHeaderName, GetRequestId(r.Context()))
			next.ServeHTTP(w, r)
		})
	}
}
