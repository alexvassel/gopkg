package email

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/severgroup-tt/gopkg-app/app"
	"github.com/severgroup-tt/gopkg-app/client/email/provider"
)

type client struct {
	metricSuccess *prometheus.Counter
	metricFailed  *prometheus.Counter
	sender        provider.ISender
	showInfo      bool
	showError     bool
}

func NewMessage(subject string) *provider.Message {
	return &provider.Message{Subject: subject}
}

func NewClient(provider provider.IProvider, fromAddress, fromName string, option ...Option) (IClient, app.PublicCloserFn, error) {
	c := client{}
	for _, o := range option {
		o(&c)
	}
	sender, closer, err := provider.Connect(fromAddress, fromName, c.showInfo, c.showError)
	if err != nil {
		return nil, closer, err
	}
	c.sender = sender
	return &c, closer, nil
}

func (c *client) Send(ctx context.Context, msg *provider.Message) error {
	if err := msg.Prepare(ctx); err != nil {
		return err
	}
	err := c.sender.Send(ctx, msg)

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
