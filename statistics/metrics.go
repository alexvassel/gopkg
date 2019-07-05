package statistics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	LastReq = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "main_last_request",
		Help: "The time of last income request",
	})

	CountError = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "main_error_count",
		Help: "The total number of request errors",
	})

	CountRequest = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "main_request_count",
		Help: "The total request count",
	})

	ResponseTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "main_response_time",
		Help: "Response time",
	})
)

var (
	reg           = prometheus.NewPedanticRegistry()
	collectorList = []prometheus.Collector{}
)

func AddCollector(collector ...prometheus.Collector) {
	collectorList = append(collectorList, collector...)
}

// Metrics prometheus metrics
func Metrics() http.Handler {
	list := collectorList
	list = append(list,
		prometheus.NewGoCollector(),
		LastReq,
		CountError,
		CountRequest,
		ResponseTime,
	)
	reg.MustRegister(list...)

	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}
