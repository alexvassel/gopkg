package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/severgroup-tt/gopkg-app/client/sentry"
	"github.com/severgroup-tt/gopkg-app/metrics"
	errors "github.com/severgroup-tt/gopkg-errors"
	"github.com/severgroup-tt/gopkg-logger"
	"google.golang.org/grpc"
	"net/http"
	"time"
)

type loggedResponseWriter struct {
	http.ResponseWriter
	status int
}

func (v *loggedResponseWriter) WriteHeader(code int) {
	v.status = code
	v.ResponseWriter.WriteHeader(code)
}

func NewLogMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			metrics.LastReq.SetToCurrentTime()
			metrics.CountRequest.Inc()

			start := time.Now().UnixNano()

			lr := &loggedResponseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(lr, r)

			reqDurationMs := (time.Now().UnixNano() - start) / int64(time.Millisecond)
			metrics.ResponseTime.Observe(float64(reqDurationMs))

			if lr.status == 500 {
				sentry.Error(errors.Internal.Err(
					r.Context(),
					fmt.Sprintf("%v %d %s %s %dms", r.RemoteAddr, lr.status, r.Method, r.URL, reqDurationMs),
				))
			}

			logger.Log(r.Context(), "%v %d %s %s %dms", r.RemoteAddr, lr.status, r.Method, r.URL, reqDurationMs)
		})
	}
}

func NewLogInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		str, _ := json.Marshal(req)
		logger.Log(ctx, "Request: %s", str)

		resp, err = handler(ctx, req)

		//str, _ = json.Marshal(resp)
		//logger.Log(ctx, "Response: %s", str)

		sentry.Error(err)

		return resp, err
	}
}
