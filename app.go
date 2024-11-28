package app

import (
	"context"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"
)

func Run(opts ...Option) error {
	return NewApp(opts...).Run()
}

type App struct {
	opts *appOptions

	plugins []Plugin
}

func NewApp(opts ...Option) *App {
	appOpts := &appOptions{
		ctx:         context.Background(),
		logger:      &defaultLogger{},
		stopTimeout: 10 * time.Second,
		stopSignals: []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
		plugins:     make([]Plugin, 0),
	}
	for _, opt := range opts {
		opt(appOpts)
	}
	return &App{
		opts: appOpts,

		plugins: make([]Plugin, 0),
	}
}

// Run 运行引擎
func (a *App) Run() (err error) {
	// 初始化上下文
	ctx := NewCtx(a.opts.ctx)
	// 注册插件
	var ok bool
	for _, p := range a.opts.plugins {
		ctx, ok, err = a.registerPlugin(ctx, p)
		switch {
		case err != nil:
			return errors.Wrapf(err, "插件%s注册失败", reflect.TypeOf(p))
		case ok:
			a.log(ctx, "插件注册成功：%s", reflect.TypeOf(p))
		default:
			a.log(ctx, "跳过插件注册：%s", reflect.TypeOf(p))
		}
	}
	// 活动周期上下文，配合其cancel方法可控制App的运行和停止
	// 活动周期：从插件注册完成后到监听到停止信号之前的生命周期定义为活动周期
	activeCtx, activeCancel := context.WithCancel(ctx)
	// 运行插件并坚挺程序的停止信号以便优雅的停止插件
	// 此处srvCtx为服务器上下文，此上下文管理的生命周期范围为服务周期（插件的Start和Stop之前的过程称为服务周期）
	eg, srvCtx := errgroup.WithContext(activeCtx)
	// swg的作用是在所有的服务自行停止时告知框架，而不是单纯的等待外部的停止命令
	var swg sync.WaitGroup
	for _, p := range a.plugins {
		p := p
		var wg sync.WaitGroup
		eg.Go(func() error {
			<-srvCtx.Done()
			// stopCtx为插件停止上下文，管理的生命周期仅为插件的停止动作周期
			// 使用目的为当插件在指定时间内没有完成停止动作，则将该上下文强行失效
			// 由于上下文只会失效，而不会强行停止这个插件，所以插件的开发者应当监听上下文的失效，并自行强制停止插件，否则可能造成框架无法停止的问题
			stopCtx, stopCancel := context.WithTimeout(ctx, a.opts.stopTimeout)
			defer stopCancel()

			return errors.WithStack(p.Stop(stopCtx))
		})
		wg.Add(1)
		swg.Add(1)
		eg.Go(func() error {
			defer swg.Done()
			wg.Done()
			return errors.WithStack(p.Start(ctx))
		})
		// 保证每个插件的Start方法都是顺序调用的
		wg.Wait()
	}
	doneChan := make(chan struct{})
	go func() {
		swg.Wait()
		doneChan <- struct{}{}
		close(doneChan)
	}()
	// 监听应用的停止信号
	c := make(chan os.Signal)
	signal.Notify(c, a.opts.stopSignals...)
	eg.Go(func() error {
		select {
		case <-doneChan:
			activeCancel()
		case <-ctx.Done():
		case <-c:
			activeCancel()
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		return errors.WithStack(err)
	}
	// 服务停止后对插件进行注销操作
	for i := len(a.plugins) - 1; i >= 0; i-- {
		err = multierr.Append(err, a.plugins[i].Deregister(ctx))
	}
	return err
}

func (a *App) registerPlugin(ctx context.Context, p Plugin) (context.Context, bool, error) {
	// 验证插件是否已经注册
	if ctx.Value(p.Key()) != nil {
		return ctx, false, nil
	}
	// 注册当前插件所依赖的插件
	var (
		ok  bool
		err error
	)
	// 自动注册依赖插件
	for _, dependPlugin := range p.Depends() {
		ctx, ok, err = a.registerPlugin(ctx, dependPlugin)
		switch {
		case ok:
			a.plugins = append(a.plugins, dependPlugin)
		case err != nil:
			return ctx, false, errors.Wrap(err, "注册依赖插件失败")
		}
	}
	// 注册插件
	if err = p.Register(ctx); err != nil {
		return ctx, false, err
	}
	// 保存插件信息并返回
	a.plugins = append(a.plugins, p)
	return context.WithValue(ctx, p.Key(), p), ok, nil
}

func (a *App) log(_ context.Context, format string, args ...any) {
	a.opts.logger.Log(format, args...)
}
