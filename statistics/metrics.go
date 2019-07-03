package statistics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	LastReq = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "slpoll_last_request",
		Help: "The time of last income request",
	})

	CountError = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "slpoll_error_count",
		Help: "The total number of request errors",
	})

	CountRequest = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "slpoll_request_count",
		Help: "The total request count",
	})

	ResponseTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "slpoll_response_time",
		Help: "Response time",
	})
)

var (
	reg = prometheus.NewPedanticRegistry()
)

// Metrics prometheus metrics
func Metrics() http.Handler {
	reg.MustRegister(
		prometheus.NewGoCollector(),
		LastReq,
		CountError,
		CountRequest,
		ResponseTime,
	)

	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}
