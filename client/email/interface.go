package email

import (
	"context"
	"github.com/severgroup-tt/gopkg-app/client/email/provider"
)

type IClient interface {
	Send(ctx context.Context, msg *provider.Message) error
}
