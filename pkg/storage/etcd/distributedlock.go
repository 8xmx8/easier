package etcd

import (
	"context"
	"fmt"

	"go.etcd.io/etcd/client/v3/concurrency"
)

const (
	lockPrefix = "/lock/%s" // /lock/<lockID>
)

func (cli *Client) Lock(ctx context.Context, key string) error {
	lock, isExist := cli.locks[key]
	if !isExist {
		sess, err := concurrency.NewSession(cli.cli)
		if err != nil {
			return err
		}
		lock = concurrency.NewMutex(sess, fmt.Sprintf(lockPrefix, key))
		cli.locks[key] = lock
	}
	if err := lock.Lock(ctx); err != nil {
		return err
	}
	// cli.logger.Info("分布式锁-lock", zap.String("key", key))
	return nil
}

func (cli *Client) UnLock(ctx context.Context, key string) error {
	lock, isExist := cli.locks[key]
	if !isExist {
		return fmt.Errorf("[%s]不存在的锁", key)
	}
	if err := lock.Unlock(ctx); err != nil {
		return err
	}
	// cli.logger.Info("分布式锁-Unlock", zap.String("key", key))
	return nil
}

func (cli *Client) DestroyLock(ctx context.Context, key string) {
	_, isExist := cli.locks[key]
	if !isExist {
		return
	}
	// cli.logger.Info("分布式锁-Destroy", zap.String("key", key))
	delete(cli.locks, key)
}
