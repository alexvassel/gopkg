package log

import (
	"github.com/severgroup-tt/gopkg-app/statistics"
	"github.com/severgroup-tt/gopkg-logger"
	"net/http"
	"time"
)

type logged_response_t struct {
	http.ResponseWriter
	status int
}

func (v *logged_response_t) WriteHeader(code int) {
	v.status = code
	v.ResponseWriter.WriteHeader(code)
}

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statistics.LastReq.SetToCurrentTime()
		statistics.CountRequest.Inc()

		start := time.Now().UnixNano()

		lr := &logged_response_t{w, http.StatusOK}
		next.ServeHTTP(lr, r)

		reqtime := (time.Now().UnixNano() - start) / int64(time.Millisecond)
		statistics.ResponseTime.Set(float64(reqtime))

		logger.Log(r.Context(), "%v %d %s %s %dms", r.RemoteAddr, lr.status, r.Method, r.URL, reqtime)
	})
}
