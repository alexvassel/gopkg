package grpc

import "time"

type Option struct {
	Service    string
	AppName    string
	AppVersion string
	Addr       string
	MaxRetry   uint
	Timeout    time.Duration
	RetryDelay time.Duration
}
