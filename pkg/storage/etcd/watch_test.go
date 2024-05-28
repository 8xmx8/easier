package etcd

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	watch_prefix = "/test/watch"
)

func TestETCDWatcher(t *testing.T) {
	ctx := context.Background()
	// new ETCD client
	cc, err := NewETCD(ctx, []string{host})
	if !assert.NoError(t, err) {
		return
	}
	// 异步执行
	cc.NewWatcher(ctx, watch_prefix, watchCallback(t))

	// 操作数据
	// 增加数据
	cc.Put(ctx, watch_prefix+"/hello", "hello")

	// 删除
	cc.Delete(ctx, watch_prefix+"/hello")

	// TTL
	cc.PutWithTTL(ctx, watch_prefix+"/hello/ttl", "this is a data with TTL", 3)
	<-time.After(5 * time.Second) // TODO: 能够监听到 `删除`时间

	// Lease
	leaseID, err := cc.CreateLeaseID(ctx, 5)
	assert.NoError(t, err)
	err = cc.PutWithLease(ctx, watch_prefix+"/hello/lease", "this is a data with lease", leaseID)
	assert.NoError(t, err)
	<-time.After(3 * time.Second)
	cc.KeepAliveOnce(ctx, leaseID)
	<-time.After(10 * time.Second)
}

func watchCallback(t *testing.T) WatcherFunc {
	return func(w *Watcher, updateEvent []*UpdateEvent, err error) {
		if err != nil {
			w.Logger.Error("watch error", zap.Error(err))
		}
		for _, ev := range updateEvent {
			switch ev.EventType {
			case UpdateEventTypePut:
				fmt.Printf("put: {%s: %s}\n", ev.Key, ev.Value)
			case UpdateEventTypeDelete:
				fmt.Printf("delete: {%s: %s}\n", ev.Key, ev.Value)
			case UpdateEventTypeUnKnown:
				fmt.Printf("UnKnown: {%s: %s}\n", ev.Key, ev.Value)
			default:
				t.Error("不是支持的事件")
			}
		}
	}
}

func TestNodeRegist(t *testing.T) {
	ctx := context.Background()
	// new ETCD client

	cc, err := NewETCD(ctx, []string{"10.10.10.93:2379"})
	if !assert.NoError(t, err) {
		return
	}
	// 异步执行
	cc.NewWatcher(ctx, "/engine", watchCallback(t))

	<-time.After(60 * time.Second)
}
