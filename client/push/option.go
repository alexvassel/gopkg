package push

import "github.com/prometheus/client_golang/prometheus"

type Option func(c *client)

func WithMetricSuccess(metric prometheus.Counter) Option {
	return func(c *client) {
		c.metricSuccess = &metric
	}
}

func WithMetricFailed(metric prometheus.Counter) Option {
	return func(c *client) {
		c.metricFailed = &metric
	}
}

func WithInfoLog() Option {
	return func(c *client) {
		c.showInfo = true
	}
}
