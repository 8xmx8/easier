package etcd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/8xmx8/easier/pkg/utils"
	etcdcli "go.etcd.io/etcd/client/v3"
	recipe "go.etcd.io/etcd/client/v3/experimental/recipes"
	"go.uber.org/zap"
)

/*
Etcd的优先级队列
*/

const (
	priorityQueuePrefix = "/priorityQueue/%s" // 优先级队列的前缀: /prioritypriorityQueue/<Key>
	priorityQueueKey    = "__/priorityQueue/%s"
)

var (
	pointIndex = len(fmt.Sprintf(priorityQueueKey, "")) // 获取已存在优先级队列名称时使用的静态标志符
)

// GetPriorityQueueList 获取指定前缀的优先级队列名称
func (cli *Client) GetPriorityQueueList(ctx context.Context, key string) ([]string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cli.timeout)
	resp, err := cli.cli.Get(timeoutCtx, fmt.Sprintf(priorityQueueKey, key), etcdcli.WithPrefix())
	cancel()
	if err != nil {
		return nil, err
	}
	queueKeySet := utils.NewStringSet()
	for _, ev := range resp.Kvs {
		key := string(ev.Key)
		end := strings.LastIndex(key, "/")
		queueKeySet.Add(key[pointIndex:end])
	}

	return queueKeySet.List(), nil
}

type PriorityQueue struct {
	logger *zap.Logger
	queue  *recipe.PriorityQueue
	etcd   *Client
	key    string
}

func NewPriorityQueue(ctx context.Context, endpoints []string, key string, ops ...OptionFunc) (*PriorityQueue, error) {
	cli, err := NewETCD(ctx, endpoints, ops...)
	if err != nil {
		return nil, err
	}
	q := recipe.NewPriorityQueue(cli.cli, fmt.Sprintf(priorityQueuePrefix, key))

	pq := &PriorityQueue{
		logger: cli.logger,
		etcd:   cli,
		queue:  q,
		key:    key,
	}
	return pq, nil
}

func NewPriorityQueueByClient(ctx context.Context, client *Client, key string) (*PriorityQueue, error) {
	q := recipe.NewPriorityQueue(client.cli, fmt.Sprintf(priorityQueuePrefix, key))
	pq := &PriorityQueue{
		logger: client.logger,
		queue:  q,
		etcd:   client,
		key:    key,
	}
	return pq, nil
}

func (p *PriorityQueue) Push(ctx context.Context, val string, pr uint16) (err error) {
	retry := 3
	for {
		err = p.queue.Enqueue(val, pr)
		if err == nil || retry < 1 {
			break
		}
		p.logger.Error("优先级队列数据写入错误", zap.Error(err), zap.Int("retry", retry))
		retry--
		<-time.After(time.Second)
	}
	return
}

// Pop 如果消息队列为空, 则该方法将阻塞
// MARK: @zcf ~~如果消息队列为空, 则该方法将阻塞1秒钟, 直到有消息可以被取出; 或为空字符串""~~
func (p *PriorityQueue) Pop(ctx context.Context) (val string, err error) {
	// MARK: @zcf Timeout太短会导返回不处数据, 推测是 超时与读取到数据竞争了
	// subctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	// go func() {
	// 	val, err = p.queue.Dequeue()
	// 	cancel()
	// }()
	// <-subctx.Done()
	return p.queue.Dequeue()
}

// Len 获取优先级队列的剩余元素
func (p *PriorityQueue) Len(ctx context.Context) (int, error) {
	resp, err := p.etcd.Get(ctx, fmt.Sprintf(priorityQueuePrefix, ""))
	if err != nil {
		return 0, err
	}
	return len(resp), nil
}

// Delete 删除这个优先级队列的所有记录
func (p *PriorityQueue) Delete(ctx context.Context) error {
	if err := p.etcd.DeleteWithPrefix(ctx, fmt.Sprintf(priorityQueuePrefix, p.key)); err != nil {
		return err
	}
	// MARK: @zcf 下列做法比较蠢, 后续考虑其必要性
	err := p.etcd.DeleteWithPrefix(ctx, fmt.Sprintf(priorityQueueKey, p.key))
	return err
}
