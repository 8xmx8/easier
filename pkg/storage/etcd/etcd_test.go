package etcd

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	host   = "10.11.12.35:2379"
	prefix = "/test"
)

func TestETCDClient(t *testing.T) {
	ctx := context.Background()
	// new ETCD client
	cc, err := NewETCD(ctx, []string{host})
	if !assert.NoError(t, err) {
		return
	}
	t.Run("put&getOne_data", func(t *testing.T) {
		key := prefix + "/put1"
		value := "this is a test data"
		err := cc.Put(ctx, key, value)
		assert.NoError(t, err)
		// 检查获取
		v, err := cc.GetOne(ctx, prefix)
		assert.NoError(t, err)
		assert.Equal(t, value, v)
	})
	t.Run("put_many&get_data", func(t *testing.T) {
		err11 := cc.Put(ctx, prefix+"/put1/put11", "this is a test data for /put1/put11")
		assert.NoError(t, err11)
		err12 := cc.Put(ctx, prefix+"/put1/put12", "this is a test data for /put1/put12")
		assert.NoError(t, err12)
		err21 := cc.Put(ctx, prefix+"/put2/put21", "this is a test data for /put2/put21")
		assert.NoError(t, err21)
		// 检查获取
		v, err := cc.Get(ctx, prefix)
		assert.NoError(t, err)
		fmt.Println(v) // TODO: 打印一下吧, assert不好验证
	})

	t.Run("put&del_data", func(t *testing.T) {
		val := "this is a test data for after del"
		key := prefix + "/putdel"
		err := cc.Put(ctx, key, val)
		assert.NoError(t, err)
		// 检查获取
		v, err := cc.GetOne(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, val, v)
		// 删除
		assert.NoError(t, cc.Delete(ctx, key))
		// 再验证
		v2, err := cc.GetOne(ctx, key)
		assert.NoError(t, err)
		assert.Empty(t, v2)
	})

	t.Run("put_with_TTL", func(t *testing.T) {
		// TODO: 与时间有关, 注意断点位置
		key := prefix + "/put_ttl"
		val := "this is a data with TTL"
		ttl := 5
		err := cc.PutWithTTL(ctx, key, val, 5)
		assert.NoError(t, err)

		time.Sleep(time.Duration(ttl-3) * time.Second)
		v1, err := cc.GetOne(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, val, v1)

		time.Sleep(time.Duration(ttl) * time.Second)
		v2, err := cc.GetOne(ctx, key)
		assert.NoError(t, err)
		assert.NotEqual(t, val, v2)
		assert.Empty(t, v2)
	})

	t.Run("lease", func(t *testing.T) {
		// TODO: 与时间有关, 注意断点位置
		ttl := 5
		key := prefix + "/put_lease"
		val := "this data with lease"
		// 创建一个测试lease
		leaseID, err := cc.CreateLeaseID(ctx, int64(ttl))
		assert.NoError(t, err)
		fmt.Println(leaseID)

		// 写入一个数据
		assert.NoError(t, cc.PutWithLease(ctx, key, val, leaseID))

		// 延时, 读取
		time.Sleep(time.Duration(2) * time.Second)
		v1, err := cc.GetOne(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, val, v1)

		// 续约
		// tl, err := cc.cli.TimeToLive(ctx, etcdcli.LeaseID(leaseID))
		// assert.NoError(t, err)
		// fmt.Println(tl.TTL)

		assert.NoError(t, cc.KeepAliveOnce(ctx, leaseID))

		// tl1, err := cc.cli.TimeToLive(ctx, etcdcli.LeaseID(leaseID))
		// assert.NoError(t, err)
		// fmt.Println(tl1.TTL)

		// 验证续约
		time.Sleep(time.Duration(ttl-1) * time.Second)
		v2, err := cc.GetOne(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, val, v2)

		// 超期验证
		time.Sleep(2 * time.Second)
		v3, err := cc.GetOne(ctx, key)
		assert.NoError(t, err)
		assert.Empty(t, v3)
	})
}

func TestWithAuth(t *testing.T) {
	user := "root"
	passwd := "123456"
	ctx := context.Background()
	// new ETCD client
	cc, err := NewETCD(ctx, []string{host}, WithBaseAuth(user, passwd))
	if !assert.NoError(t, err) {
		return
	}
	res, err := cc.Get(ctx, "")
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println(res)
}

func TestCompact(t *testing.T) {
	ctx := context.Background()
	// new ETCD client
	cc, err := NewETCD(ctx, []string{host})
	if !assert.NoError(t, err) {
		return
	}
	t.Run("Compact", func(t *testing.T) {
		cc.Compact(ctx, 6, false)
	})
	t.Run("Compact", func(t *testing.T) {
		cc.Compact(ctx, 6, true)
	})
}
