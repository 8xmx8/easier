package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/8xmx8/easier/pkg/logger"
	"github.com/go-redis/redis/v7"
	"github.com/jinzhu/copier"
)

// Cluster 集群模式
type Cluster struct {
	cluster *redis.ClusterClient
	logg    logger.Logger
}

type OptionFuncForCluster func(*redis.ClusterOptions)

// WithPasswd 配置密码
func ClusterWithPasswd(passwd string) OptionFuncForCluster {
	return func(o *redis.ClusterOptions) {
		o.Password = passwd
	}
}

// WithAuth 配置鉴权
func ClusterWithAuth(user, passwd string) OptionFuncForCluster {
	return func(o *redis.ClusterOptions) {
		o.Username = user
		o.Password = passwd
	}
}

func NewCluster(ctx context.Context, addr []string, ops ...OptionFuncForCluster) (Redis, error) {
	opt := &redis.ClusterOptions{
		Addrs:         addr,
		RouteRandomly: true,
	}
	for _, op := range ops {
		op(opt)
	}
	return &Cluster{
		cluster: redis.NewClusterClient(opt).WithContext(ctx),
		logg:    logger.DefaultLogger(),
	}, nil
}

/********************************* 通用接口 **************************************/
// IsExist 判断key是否存在
func (c *Cluster) IsExist(ctx context.Context, key ...string) bool {
	return c.cluster.Exists(key...).Val() > 0
}

// Del 删除key
func (c *Cluster) Del(ctx context.Context, key ...string) error {
	return c.cluster.Del(key...).Err()
}

// SetExpire 设置key的ttl
func (c *Cluster) SetExpire(ctx context.Context, key string, ttl time.Duration) error {
	return c.cluster.Expire(key, ttl).Err()
}

// GetExpire 获取ttl
func (c *Cluster) GetExpire(ctx context.Context, key string) (time.Duration, error) {
	return c.cluster.TTL(key).Result()
}

// FlushDB 删除数据
// isAll 是否使用FlushAll
func (c *Cluster) FlushDB(ctx context.Context, isAll bool) error {
	if isAll {
		return c.cluster.FlushAll().Err()
	}
	return c.cluster.FlushDB().Err()
}

// GetMixed 获取到key对应的value
// mixed redis's basic type
// TODO: @zcf 根据实际应用,考虑在功能前加入对"key是否存在进行判断"
func (c *Cluster) GetMixed(ctx context.Context, key string, value interface{}) error {
	_type, err := c.getType(ctx, key)
	if err != nil {
		return err
	}
	switch _type {
	case "string": // nolint
		// "hello"
		return c.cluster.Get(key).Scan(value)
	case "list": // nolint
		// 与RPUSH 组成先进先出列表 []string
		return c.cluster.LRange(key, 0, -1).ScanSlice(value)
	case "set": // nolint
		// 元素唯一的string类型的无序集合 []string
		return c.cluster.SMembers(key).ScanSlice(value)
	case "zset": // nolint
		// 元素唯一的string类型的有序集合 []string
		return c.cluster.ZRevRange(key, 0, -1).ScanSlice(value)
	case "hash": // nolint
		// object
		h := c.cluster.HGetAll(key)
		if h.Err() != nil {
			return h.Err()
		}
		hb, err := json.Marshal(h.Val())
		if err != nil {
			return err
		}
		return json.Unmarshal(hb, value)
	default:
		// return fmt.Errorf("%s 不存在", key)
		return nil
	}
}

// getType 获取key对应数据的数据类型
func (c *Cluster) getType(ctx context.Context, key string) (string, error) {
	return c.cluster.Type(key).Result()
}

// ScanKey 扫描字符串值
func (c *Cluster) ScanKey(ctx context.Context, match string) chan string {
	batchSize := 10
	out := make(chan string, batchSize)
	go func() {
		defer close(out)
		var (
			cursor uint64
			err    error
		)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var keys []string
				keys, cursor, err = c.cluster.Scan(cursor, match, int64(batchSize)).Result()
				if err != nil {
					c.logg.Error(logger.ErrorCache, "redis scan error", logger.MakeField("match", match), logger.ErrorField(err))
					return
				}
				for _, item := range keys {
					out <- item
				}
				if cursor == 0 {
					return
				}
			}
		}
	}()
	return out
}

/********************************* string接口 **************************************/
// Set 设置key数据(不含TTL)
func (c *Cluster) SetStr(ctx context.Context, key, value string) error {
	return c.cluster.Set(key, value, 0).Err()
}

// SetStrTTL 设置key数据(含TTL)
func (c *Cluster) SetStrTTL(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.cluster.Set(key, value, ttl).Err()
}

// SetNX 设置分布式锁
func (c *Cluster) SetNX(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.cluster.WithContext(ctx).SetNX(key, value, ttl).Err()
}

/*********************************** hash接口 ****************************************/

// SetHash设置hash数据
//
//	HMSet("myhash", map[string]interface{}{"key1": "value1", "key2": "value2"})
//
// TODO: @zcf redis 的Hash不支持 层级结构
func (c *Cluster) SetHash(ctx context.Context, key string, value map[string]interface{}) error {
	return c.cluster.HMSet(key, value).Err()
}

// GetHashField 获取执行hash的指定field数据
// TODO: @zcf 当field对应的数据为非string数据类型时, 需要对该方法进行改造(例如: 返回[]byte); 如果要改造,需要对redis interface同步修改
func (c *Cluster) GetHashField(ctx context.Context, key, field string) (string, error) {
	return c.cluster.HGet(key, field).Result()
}

/*********************************** list接口 ****************************************/
// TODO: @zcf 针对list类型的头部（左边）和尾部（右边）的操作暂时不做多余实现.
// 当前默认, list添加数据从尾部（右边）添加, 从头部（左边）开始获取

// PushList 向key对应的list中从尾部(右边)追加数据
func (c *Cluster) PushList(ctx context.Context, key string, values ...interface{}) error {
	return c.cluster.RPush(key, values...).Err()
}

// LenList 获取指定列表长度
func (c *Cluster) LenList(ctx context.Context, key string) (int64, error) {
	return c.cluster.LLen(key).Result()
}

// PopList 从头部（左边）开始获取并删除
func (c *Cluster) PopList(ctx context.Context, key string, value interface{}) error {
	return c.cluster.LPop(key).Scan(value)
}

/*********************************** set接口 ****************************************/
// TODO: @zcf 当前进提供最基本的操作, 暂时不支持集合运算

// AddSet 向key集合中添加成员
func (c *Cluster) AddSet(ctx context.Context, key string, values ...interface{}) error {
	return c.cluster.SAdd(key, values...).Err()
}

// CheckSetMember 检查成员是否在集合内
func (c *Cluster) CheckSetMember(ctx context.Context, key string, value interface{}) (bool, error) {
	return c.cluster.SIsMember(key, value).Result()
}

// RemSetEle 移除key集合中的元素
func (c *Cluster) RemSetEle(ctx context.Context, key string, values ...interface{}) error {
	return c.cluster.SRem(key, values...).Err()
}

/*********************************** 有序集合接口 ****************************************/

// AddZSet 向key有序集合添加成员
// members map[成员数据]分数
func (c *Cluster) AddZSet(ctx context.Context, key string, members ...*ZSetMember) error {
	zms := make([]*redis.Z, 0, len(members))
	for _, m := range members {
		zms = append(zms, &redis.Z{Score: m.Score, Member: m.Member})
	}
	return c.cluster.ZAdd(key, zms...).Err()
}

// CardZSet 获取有序集合的成员数
func (c *Cluster) CardZSet(ctx context.Context, key string) (int64, error) {
	return c.cluster.ZCard(key).Result()
}

// MembersWithScoreZSet 从高到低获取有序集合的成员及分数
func (c *Cluster) MembersWithScoreZSet(ctx context.Context, key string) ([]*ZSetMember, error) {
	zres, err := c.cluster.ZRevRangeWithScores(key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	values := make([]*ZSetMember, 0, len(zres))
	err = copier.Copy(&values, &zres)
	return values, err
}

// RemMembersZSet 移除key有序集合中的members
func (c *Cluster) RemMembersZSet(ctx context.Context, key string, members ...string) error {
	return c.cluster.ZRem(key, members).Err()
}

/*********************************** pubsub接口 ****************************************/
func (c *Cluster) Publish(ctx context.Context, key string, value interface{}) error {
	return c.cluster.Publish(key, value).Err()
}

func (c *Cluster) Subscribe(ctx context.Context, channels ...string) (<-chan *Message, error) {
	pubsub := c.cluster.Subscribe(channels...)
	_, err := pubsub.Receive()
	if err != nil {
		return nil, err
	}
	ch := make(chan *Message, 1)
	go func() {
		defer close(ch)
		for msg := range pubsub.Channel() {
			ch <- &Message{
				Channel: msg.Channel,
				Payload: msg.Payload,
			}
		}
	}()
	return ch, nil
}
