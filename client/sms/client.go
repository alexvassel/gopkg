package sms

import (
	"context"
	"github.com/severgroup-tt/gopkg-app/client/sms/provider"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/severgroup-tt/gopkg-app/app"
)

type client struct {
	metricSuccess *prometheus.Counter
	metricFailed  *prometheus.Counter
	sender        provider.ISender
	showInfo      bool
	showError     bool
}

func NewClient(provider provider.IProvider, option ...Option) (IClient, app.PublicCloserFn, error) {
	c := &client{}
	for _, o := range option {
		o(c)
	}
	sender, closer, err := provider.Connect(c.showInfo, c.showError)
	if err != nil {
		return nil, closer, err
	}
	c.sender = sender
	return c, closer, nil
}

func (c client) Send(ctx context.Context, phone int64, message string) error {
	err := c.sender.Send(ctx, phone, message)
	if err == nil {
		if c.metricSuccess != nil {
			(*c.metricSuccess).Inc()
		}
	} else {
		if c.metricFailed != nil {
			(*c.metricFailed).Inc()
		}
	}
	return err
}
