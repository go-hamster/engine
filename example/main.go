package main

import (
	"context"
	"log"
	"time"

	"github.com/go-hamster/engine"
)

func main() {
	err := app.Run(
		app.WithStopTimeout(time.Second*5),
		app.WithPlugins(&DebugPlugin{}),
	)
	if err != nil {
		panic(err)
	}
}

type DebugPlugin struct {
	app.PluginAdapter
}

type debugPluginCtxKey struct{}

func (p *DebugPlugin) Key() any {
	return debugPluginCtxKey{}
}

func (p *DebugPlugin) Register(_ context.Context) error {
	log.Println("register")
	return nil
}

func (p *DebugPlugin) Start(_ context.Context) error {
	time.Sleep(time.Second * 3)
	log.Println("start")
	return nil
}
