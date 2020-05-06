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

const loggerLevel = 500

type loggedResponseWriter struct {
	http.ResponseWriter
	status int
	error string
}

func (v *loggedResponseWriter) WriteHeader(code int) {
	v.status = code
	v.ResponseWriter.WriteHeader(code)
}

func (v *loggedResponseWriter) Write(bytes []byte) (int, error) {
	if v.status >= loggerLevel {
		v.error += string(bytes)
	}
	return v.ResponseWriter.Write(bytes)
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

			if lr.status >= loggerLevel {
				logger.Error(r.Context(), lr.error)
				sentry.Error(errors.Internal.Err(
					r.Context(),
					fmt.Sprintf("%v %d %s %s %s %dms", r.RemoteAddr, lr.status, r.Method, r.URL, lr.error, reqDurationMs),
				))
			}

			logger.Info(r.Context(), "%v %d %s %s %dms", r.RemoteAddr, lr.status, r.Method, r.URL, reqDurationMs)
		})
	}
}

func NewLogInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		msgFormat := "method: %s, request: %s"
		str, _ := json.Marshal(req)
		if len(str) > 500 {
			str = str[:500]
			msgFormat += " ..."
		}
		msgParam := []interface{}{info.FullMethod, str}
		kvList := logExtraFromContext(ctx)
		if len(kvList)&1 == 1 {
			kvList = append(kvList, "?")
		}
		for i := 0; i < len(kvList); i += 2 {
			msgFormat = "%s: %v, " + msgFormat
			msgParam = append([]interface{}{kvList[i], kvList[i+1]}, msgParam...)
		}
		logger.Info(ctx, msgFormat, msgParam...)

		resp, err = handler(ctx, req)

		//str, _ = json.Marshal(resp)
		//logger.Info(ctx, "Response: %s", str)

		sentry.Error(err)

		return resp, err
	}
}
