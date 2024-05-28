package etcd

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/8xmx8/easier/pkg/logger"
	"github.com/8xmx8/easier/pkg/utils"
	etcd3 "go.etcd.io/etcd/client/v3"
)

func NewProviderKey(preifx, host string, debug bool) (k, v string) {
	var key = preifx + "/" + host
	if debug {
		key = "/debug" + key
	}
	// MARK: @zcf 这里可以增加一个"获取本机全部可用网卡的地址"的逻辑, 用于直接获取本机的地址
	value := utils.SetProtocol(host, utils.HTTP)
	return key, value
}

// Provider 注册一个Key，并且随时保持心跳，对象同样有Start和Stop函数.
type Provider struct {
	client   *Client
	logg     logger.Logger
	cancel   func()
	ctx      context.Context
	callback ProviderFunc
	key      string
	value    string
	ttl      int64
	wait     sync.WaitGroup
}

type ProviderFunc func(p *Provider, err error)

// NewProvider 创建一个Provider.
func NewProvider(ctx context.Context, c *Client, key, value string, ttl int64, f ProviderFunc) error {
	if c == nil || c.cli == nil {
		return errors.New("not connected etcd")
	}
	if f == nil {
		return errors.New("invalid param")
	}
	p := &Provider{
		key:      key,
		value:    value,
		ttl:      ttl,
		client:   c,
		callback: f,
	}
	p.ctx, p.cancel = context.WithCancel(ctx)
	return p.Start()
}

func (p *Provider) grant() (<-chan *etcd3.LeaseKeepAliveResponse, error) {
	resp, err := p.client.cli.Grant(p.ctx, p.ttl)
	if err != nil {
		return nil, err
	}
	_, err = p.client.cli.Put(p.ctx, p.key, p.value, etcd3.WithLease(resp.ID))
	if err != nil {
		return nil, err
	}
	return p.client.cli.KeepAlive(p.ctx, resp.ID)
}

// Start 保持与ETCD之间的连接.
func (p *Provider) Start() error {
	tick := time.NewTicker(time.Second)
	p.wait.Add(1)
	go func(c context.Context) {
		defer p.wait.Done()
		for {
			select {
			case <-tick.C:
				ch, err := p.grant()
				if err != nil {
					p.logg.Error(logger.ErrETCDException, "etcd操作错误", logger.ErrorField(err))
					continue
				}
				// NOTICE 此循环是进程保活使用的
				for disconnect := false; !disconnect; {
					select {
					case _, ok := <-ch:
						if !ok {
							disconnect = true
							p.callback(p, fmt.Errorf("provider is disconnect"))
							break
						}
					case <-c.Done(): // 退出线程:
						return
					}
				}
			case <-c.Done(): // 退出线程:
				return
			}
		}
	}(p.ctx)
	return nil
}

func (p *Provider) Stop() {
	p.cancel()
	p.wait.Wait()
}
