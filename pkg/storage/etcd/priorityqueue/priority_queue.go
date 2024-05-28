package priorityqueue

import (
	"context"

	spb "go.etcd.io/etcd/api/v3/mvccpb"
	v3 "go.etcd.io/etcd/client/v3"
	recipe "go.etcd.io/etcd/client/v3/experimental/recipes"
)

type PriorityQueue struct {
	RecipeQueue *recipe.PriorityQueue
	client      *v3.Client
	ctx         context.Context
	key         string
}

func NewPriorityQueue(client *v3.Client, key string) *PriorityQueue {
	return &PriorityQueue{
		RecipeQueue: recipe.NewPriorityQueue(client, key),
		client:      client,
		ctx:         context.TODO(),
		key:         key + "/",
	}
}

// 无阻塞出列
func (p *PriorityQueue) BDequeue() (string, error) {
	resp, err := p.client.Get(p.ctx, p.key, v3.WithFirstKey()...)
	if err != nil {
		return "", err
	}

	kv, err := claimFirstKey(p.client, resp.Kvs)
	if err != nil {
		return "", err
	} else if kv != nil {
		return string(kv.Value), nil
	} else if resp.More {
		// missed some items, retry to read in more
		return p.BDequeue()
	} else {
		return "", err
	}
}

// 删除优先级队列
func (p *PriorityQueue) Delete() error {
	_, err := p.client.Delete(p.ctx, p.key)
	if err != nil {
		return err
	}
	return nil
}

// deleteRevKey deletes a key by revision, returning false if key is missing
func deleteRevKey(kv v3.KV, key string, rev int64) (bool, error) {
	cmp := v3.Compare(v3.ModRevision(key), "=", rev)
	req := v3.OpDelete(key)
	txnresp, err := kv.Txn(context.TODO()).If(cmp).Then(req).Commit()
	if err != nil {
		return false, err
	} else if !txnresp.Succeeded {
		return false, nil
	}
	return true, nil
}

func claimFirstKey(kv v3.KV, kvs []*spb.KeyValue) (*spb.KeyValue, error) {
	for _, k := range kvs {
		ok, err := deleteRevKey(kv, string(k.Key), k.ModRevision)
		if err != nil {
			return nil, err
		} else if ok {
			return k, nil
		}
	}
	return nil, nil
}
