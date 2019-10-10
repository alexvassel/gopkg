package grpc

import "time"

const (
	DefaultMaxRetry   = uint(5)
	DefaultRetryDelay = 500 * time.Millisecond
)

type Option struct {
	Service    string
	AppName    string
	AppVersion string
	Addr       string
	MaxRetry   uint
	Timeout    time.Duration
	RetryDelay time.Duration
}
