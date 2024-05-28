package utils

/* 并发控制: 协程池 */

import (
	"fmt"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

type ConcurrentPool struct {
	p    *ants.Pool
	lock sync.Mutex
	wg   sync.WaitGroup
}

// WithExpiryDuration 设置清理协程的时间间隔
var WithExpiryDuration = ants.WithExpiryDuration

// WithLogger 自定义logger
var WithLogger = ants.WithLogger

// WithPanicHandler 自定义panic处理逻辑
var WithPanicHandler = ants.WithPanicHandler

// MARK: @zcf: 提供了以下选项, 但不推荐使用
// WithNonblocking 当没有可用的工作者时, 池是否将返回nil
// var WithNonblocking = ants.WithNonblocking

// WithMaxBlockingTasks 设置当达到池的容量时被阻塞的goroutines的最大数量。
// var WithMaxBlockingTasks = ants.WithMaxBlockingTasks

// WithPreAlloc 是否应该为工作者分配内存(malloc)
// var WithPreAlloc = ants.WithPreAlloc
// nolint
func defaultPanicHandler(i interface{}) {
	fmt.Printf("协程池处理异常: %v\n", i)
}

// NewPool 实例化预分配协程池
func NewPool(num int, opts ...ants.Option) (*ConcurrentPool, error) {
	if num < 1 {
		return nil, fmt.Errorf("pool size should be greater than 1, not %d", num)
	}
	ops := []ants.Option{ants.WithOptions(ants.Options{
		ExpiryDuration:   5 * time.Second,
		PreAlloc:         true,
		MaxBlockingTasks: num,
		Nonblocking:      false, // 越界阻塞
		PanicHandler:     defaultPanicHandler,
		// Logger:           logx.WithContext(context.Background()),
	})}
	ops = append(ops, opts...)
	p, err := ants.NewPool(num, ops...)
	if err != nil {
		return nil, err
	}
	return &ConcurrentPool{
		p:    p,
		lock: sync.Mutex{},
		wg:   sync.WaitGroup{},
	}, nil
}

// Submit 注册一个任务
func (c *ConcurrentPool) Submit(f func()) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.wg.Add(1)
	return c.p.Submit(func() {
		defer c.wg.Done()
		f()
	})
}

// Wait 等待阻塞
func (c *ConcurrentPool) Wait() {
	c.wg.Wait()
}

// WaitNum 等待执行的任务数
func (c *ConcurrentPool) WaitAndRunNum() (int, int) {
	return c.p.Waiting(), c.p.Running()
}

// Release 释放这个协程池
func (c *ConcurrentPool) Release() {
	if c.p != nil {
		c.p.Release()
	}
	c.p = nil
}
