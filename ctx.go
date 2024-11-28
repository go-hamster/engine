package app

import (
	"context"
	"time"
)

type Ctx struct {
	ctx context.Context
}

func (c *Ctx) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

func (c *Ctx) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *Ctx) Err() error {
	return c.ctx.Err()
}

func (c *Ctx) Value(key any) any {
	return c.ctx.Value(key)
}

var _ context.Context = (*Ctx)(nil)

func NewCtx(ctx context.Context) context.Context {
	return &Ctx{ctx: ctx}
}
