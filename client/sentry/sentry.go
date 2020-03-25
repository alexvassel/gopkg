package sentry

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	errors "github.com/severgroup-tt/gopkg-errors"
	logger "github.com/severgroup-tt/gopkg-logger"
	"time"
)

var instance *sentryConfig

type sentryConfig struct {
	flushTimeout time.Duration
}

func Init(project, token string, flushTimeout time.Duration) error {
	dsn := fmt.Sprintf("https://%s@sentry.io/%s", token, project)
	err := sentry.Init(sentry.ClientOptions{Dsn: dsn})
	if err != nil {
		logger.Info(logger.App, "Can't connect to Sentry dsn "+dsn+": "+err.Error())
		return errors.Internal.ErrWrap(context.Background(), "Sentry initialization failed", err).
			WithLogKV("project", project, "token", token, "dsn", dsn)
	}
	logger.Info(logger.App, "Connect to Sentry project "+project)
	instance = &sentryConfig{flushTimeout: flushTimeout}
	return nil
}

func Error(err error) {
	if instance == nil || sentry.CurrentHub() == nil {
		return
	}
	if errors.IsInternal(err) {
		sentry.CaptureException(err)
		sentry.Flush(instance.flushTimeout)
	}
}

func Panic(err interface{}) {
	if instance == nil {
		return
	}
	if hub := sentry.CurrentHub(); hub != nil {
		hub.Recover(fmt.Sprintf("%#v", err))
		sentry.Flush(instance.flushTimeout)
	}
}
