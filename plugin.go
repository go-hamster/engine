package app

import (
	"context"
)

type Plugin interface {
	// Key 插件的key，需要保证在整个引擎中唯一
	Key() any
	// Register 引擎在注册插件时会执行该方法
	Register(ctx context.Context) error
	// Start 引擎完成准备后调用插件的 Start 方法进行运行插件内的服务（该方法会在协程中执行）
	Start(ctx context.Context) error
	// Stop 档引擎接受到停止的信号，会调用插件的 Stop 方法来告知插件应当停止内部服务（该方法会在携程中执行）
	// PS：上下文中将会携带超时信号，建议在 Stop 方法内部针对超时进行处理，以防止引擎无法正常停止
	Stop(ctx context.Context) error
	// Deregister 引擎在注销插件时会调用该方法，此时插件可以做一些资源释放的操作
	Deregister(ctx context.Context) error
	// Depends 告知引擎这个插件依赖哪些插件，如果所依赖的插件没有被注册，则应用会尝试自动注册
	Depends() []Plugin
}

type PluginAdapter struct{}

func (a PluginAdapter) Register(_ context.Context) error {
	return nil
}

func (a PluginAdapter) Start(_ context.Context) error {
	return nil
}

func (a PluginAdapter) Stop(_ context.Context) error {
	return nil
}

func (a PluginAdapter) Deregister(_ context.Context) error {
	return nil
}

func (a PluginAdapter) Depends() []Plugin {
	return []Plugin{}
}
