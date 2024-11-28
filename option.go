package app

import (
	"context"
	"os"
	"time"
)

type appOptions struct {
	// 跟上下文，App内部的一切上下文均由此派生
	ctx context.Context

	logger Logger

	stopTimeout time.Duration
	stopSignals []os.Signal

	plugins []Plugin
}

type Option func(opts *appOptions)

func WithCtx(ctx context.Context) Option {
	return func(opts *appOptions) {
		opts.ctx = ctx
	}
}

func WithLogger(logger Logger) Option {
	return func(opts *appOptions) {
		opts.logger = logger
	}
}

func WithStopTimeout(timeout time.Duration) Option {
	return func(opts *appOptions) {
		opts.stopTimeout = timeout
	}
}

func WithStopSignal(signals ...os.Signal) Option {
	return func(opts *appOptions) {
		opts.stopSignals = signals
	}
}

func WithPlugins(plugins ...Plugin) Option {
	return func(opts *appOptions) {
		opts.plugins = append(opts.plugins, plugins...)
	}
}
