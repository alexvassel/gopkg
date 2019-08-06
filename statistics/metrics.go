package statistics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	reg           = prometheus.NewPedanticRegistry()
	collectorList = []prometheus.Collector{}

	LastReq      prometheus.Gauge
	CountError   prometheus.Counter
	CountRequest prometheus.Counter
	ResponseTime prometheus.Gauge
)

func AddBasicCollector(prefix string) {
	LastReq = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: prefix + "_last_request",
		Help: "The time of last income request",
	})

	CountError = prometheus.NewCounter(prometheus.CounterOpts{
		Name: prefix + "_error_count",
		Help: "The total number of request errors",
	})

	CountRequest = prometheus.NewCounter(prometheus.CounterOpts{
		Name: prefix + "_request_count",
		Help: "The total request count",
	})

	ResponseTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: prefix + "_response_time",
		Help: "Response time",
	})

	collectorList = append(collectorList,
		prometheus.NewGoCollector(),
		LastReq,
		CountError,
		CountRequest,
		ResponseTime,
	)
}

func AddCollector(collector ...prometheus.Collector) {
	collectorList = append(collectorList, collector...)
}

// Metrics prometheus metrics
func Metrics() http.Handler {
	reg.MustRegister(collectorList...)
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}
