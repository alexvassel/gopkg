package sms

import (
	"context"
)

type IClient interface {
	Send(ctx context.Context, phone int64, message string) error
}
