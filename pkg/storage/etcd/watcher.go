package etcd

import (
	"context"
	"fmt"
	"time"

	"github.com/8xmx8/easier/pkg/utils"
	etcdcli "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type EventType int

const (
	// ETCD 操作类型
	UpdateEventTypePut EventType = iota + 1
	UpdateEventTypeDelete
	UpdateEventTypeUnKnown
)

type UpdateEvent struct {
	Key       string
	Value     string
	PreKey    string
	PreValue  string
	EventType EventType
}

func (ue UpdateEvent) String() string {
	return fmt.Sprintf("{Key:%s, Value:%s, EventType:%d}", ue.Key, ue.Value, ue.EventType)
}

// Watcher 对象，用于监听一个Key的变化对象，这个对象有Start、Stop参数.
type Watcher struct { // nolint
	client   *Client
	callback WatcherFunc
	prefix   string
	cancel   func()
	ctx      context.Context
	Logger   *zap.Logger
	pools    *utils.ConcurrentPool
}

// WatcherFunc 来自于Watcher的回调，通过这个回调来进行通知.
type WatcherFunc func(w *Watcher, updateEvent []*UpdateEvent, err error)

func (w *Watcher) Start(ctx context.Context) error {
	go w.runWatcher(ctx)
	return nil
}

func (w *Watcher) runWatcher(ctx context.Context) {
	for {
		subctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
		w.handlerWatch(subctx, w.prefix, w.callback)
		<-subctx.Done()
		cancel() // 尝试再关闭一下
		w.Logger.Info("watcher has new", zap.String("prefix", w.prefix))
	}
}
func (w *Watcher) handlerWatch(ctx context.Context, prefix string, callback WatcherFunc) {
	wc := w.client.cli.Watch(ctx, prefix, etcdcli.WithPrefix(), etcdcli.WithPrevKV())
	for {
		select {
		case resp, ok := <-wc:
			if !ok {
				callback(w, nil, nil)
				time.Sleep(time.Second) // 缓冲一下
				w.Logger.Error("etcd监听通道关闭, 5秒钟后重试")
				break
			}
			for _, ev := range resp.Events {
				var eventType EventType
				switch ev.Type {
				case etcdcli.EventTypePut:
					eventType = UpdateEventTypePut
				case etcdcli.EventTypeDelete:
					eventType = UpdateEventTypeDelete
				default:
					eventType = UpdateEventTypeUnKnown
				}
				data := &UpdateEvent{
					EventType: eventType,
					Key:       string(ev.Kv.Key),
					Value:     string(ev.Kv.Value),
				}
				if ev.PrevKv != nil {
					data.PreKey = string(ev.PrevKv.Key)
					data.PreValue = string(ev.PrevKv.Value)
				}
				if err := w.pools.Submit(func() {
					callback(w, []*UpdateEvent{data}, nil)
				}); err != nil {
					w.Logger.Error("watcher 协程池任务创建错误", zap.Error(err))
				}
			}
		case <-ctx.Done():
			w.Logger.Info("watcher has stopped", zap.String("prefix", w.prefix))
			return
		}
	}
}

// Stop 停止监听.
func (w *Watcher) Stop() {
	w.cancel()
}
