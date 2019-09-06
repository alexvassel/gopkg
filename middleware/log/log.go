package log

import (
	"context"
	"encoding/json"
	"github.com/severgroup-tt/gopkg-app/statistics"
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
			statistics.LastReq.SetToCurrentTime()
			statistics.CountRequest.Inc()

			start := time.Now().UnixNano()

			lr := &loggedResponseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(lr, r)

			reqDuration := (time.Now().UnixNano() - start) / int64(time.Millisecond)
			statistics.ResponseTime.Set(float64(reqDuration))

			logger.Log(r.Context(), "%v %d %s %s %dms", r.RemoteAddr, lr.status, r.Method, r.URL, reqDuration)
		})
	}
}

func NewLogInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		str, _ := json.Marshal(resp)
		logger.Log(ctx, "Request: %s", str)

		resp, err = handler(ctx, req)

		str, _ = json.Marshal(resp)
		logger.Log(ctx, "Response: %s", str)

		return resp, err
	}
}
