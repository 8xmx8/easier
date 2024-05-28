package etcd

import (
	"context"
	"errors"
	"time"

	"go.etcd.io/etcd/client/v3/concurrency"
	"go.uber.org/zap"
)

// ElectionFunc 选举成功之后会回调这个函数，只会回调一次.
type ElectionFunc func(ctx context.Context, e *Election, isLeader bool, err error)

type ElectionConfig struct {
	Timeout int `json:"timeout"`
}

// Election 用来进行选举，如果成功，就会通过回调函数返回选举成功，默认情况下，当前节点不是主节点.
type Election struct { // nolint
	key      string
	nodeName string
	conf     *ElectionConfig
	cs       *concurrency.Session
	ele      *concurrency.Election
	callback ElectionFunc
	logger   *zap.Logger
	cancel   func()
}

func (e *Election) Start(ctx context.Context) (err error) {
	// 重置选举状态
	e.callback(ctx, e, false, errors.New("重新选举"))
	e.monitor(ctx)

	return nil
}

// monitor leader监视器
func (e *Election) monitor(ctx context.Context) {
	for {
		// 判断是否有leader
		isloader, noLeader, err := e.leader(ctx)
		if err != nil {
			e.logger.Warn("leader err", zap.Error(err))
			e.callback(ctx, e, false, err) // 不是leader的回调
			continue
		}
		if noLeader { // 没有leader,发起选举请求
			e.callback(ctx, e, false, nil)
			if err := e.elect(ctx); err != nil {
				e.logger.Warn("elect err", zap.Error(err))
			}
			continue
		}
		if isloader {
			e.callback(ctx, e, true, nil)
		} else {
			e.callback(ctx, e, false, nil)
		}
	}
}

// elect 选举
func (e *Election) elect(ctx context.Context) error {
	// 添加超时时间
	tctx, cancel := context.WithTimeout(ctx, time.Duration(e.conf.Timeout)*time.Second)
	defer cancel()
	return e.ele.Campaign(tctx, e.key)
}

// 查询leader的信息
func (e *Election) leader(ctx context.Context) (isLeader, noLeader bool, err error) {
	res, err := e.ele.Leader(ctx)
	if err != nil {
		if err == concurrency.ErrElectionNoLeader { // 没有leader
			return false, true, nil
		}
		return false, false, err // 单纯报错
	}
	if string(res.Kvs[0].Value) == e.nodeName {
		return true, false, nil
	}
	return false, false, nil
}

// Stop 停止进行选举.
func (e *Election) Destroy() {
	if e.cs != nil {
		e.cs.Close()
	}
	e.cs = nil
}
