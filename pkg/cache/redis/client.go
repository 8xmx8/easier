package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	"time"

	"github.com/8xmx8/easier/pkg/logger"
	"github.com/jinzhu/copier"
)

// Client 单机模式
type Client struct {
	client *redis.Client
	logg   logger.Logger
}

type OptionFunc func(*redis.Options)

// WithPasswd 配置密码
func WithPasswd(passwd string) OptionFunc {
	return func(o *redis.Options) {
		o.Password = passwd
	}
}

// WithAuth 配置鉴权
func WithAuth(user, passwd string) OptionFunc {
	return func(o *redis.Options) {
		o.Username = user
		o.Password = passwd
	}
}

func NewClient(ctx context.Context, addr string, db int, ops ...OptionFunc) (Redis, error) {
	if db > 15 {
		return nil, fmt.Errorf("client mode Only libraries 0 to 15 are supported, currently as %d", db)
	}
	opt := &redis.Options{
		Addr: addr,
		DB:   db,
	}
	for _, op := range ops {
		op(opt)
	}

	return &Client{
		client: redis.NewClient(opt).WithContext(ctx),
		logg:   logger.DefaultLogger(),
	}, nil
}

/********************************* 通用接口 **************************************/
// IsExist 判断key是否存在
func (c *Client) IsExist(ctx context.Context, key ...string) bool {
	return c.client.Exists(key...).Val() > 0
}

// Del 删除key
func (c *Client) Del(ctx context.Context, key ...string) error {
	return c.client.Del(key...).Err()
}

// SetExpire 设置key的ttl
func (c *Client) SetExpire(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(key, ttl).Err()
}

// GetExpire 获取ttl
func (c *Client) GetExpire(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(key).Result()
}

// FlushDB 删除数据
// isAll 是否使用FlushAll
func (c *Client) FlushDB(ctx context.Context, isAll bool) error {
	if isAll {
		return c.client.FlushAll().Err()
	}
	return c.client.FlushDB().Err()
}

// GetMixed 获取到key对应的value
// mixed redis's basic type
// TODO: @zcf 根据实际应用,考虑在功能前加入对"key是否存在进行判断"
func (c *Client) GetMixed(ctx context.Context, key string, value interface{}) error {
	_type, err := c.getType(ctx, key)
	if err != nil {
		return err
	}
	switch _type {
	case "string": //nolint
		// "hello"
		return c.client.Get(key).Scan(value)
	case "list": //nolint
		// 与RPUSH 组成先进先出列表 []string
		return c.client.LRange(key, 0, -1).ScanSlice(value)
	case "set": //nolint
		// 元素唯一的string类型的无序集合 []string
		return c.client.SMembers(key).ScanSlice(value)
	case "zset": //nolint
		// 元素唯一的string类型的有序集合 []string
		return c.client.ZRevRange(key, 0, -1).ScanSlice(value)
	case "hash": //nolint
		// object
		h := c.client.HGetAll(key)
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

// ScanKey 扫描字符串值
func (c *Client) ScanKey(ctx context.Context, match string) chan string {
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
				keys, cursor, err = c.client.Scan(cursor, match, int64(batchSize)).Result()
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

// getType 获取key对应数据的数据类型
func (c *Client) getType(ctx context.Context, key string) (string, error) {
	return c.client.Type(key).Result()
}

/********************************* string接口 **************************************/
// Set 设置key数据(不含TTL)
func (c *Client) SetStr(ctx context.Context, key, value string) error {
	return c.client.Set(key, value, 0).Err()
}

// SetStrTTL 设置key数据(含TTL)
func (c *Client) SetStrTTL(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.client.Set(key, value, ttl).Err()
}

// SetNX 设置分布式锁
func (c *Client) SetNX(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.client.WithContext(ctx).SetNX(key, value, ttl).Err()
}

/*********************************** hash接口 ****************************************/

// SetHash设置hash数据
//
//	HMSet("myhash", map[string]interface{}{"key1": "value1", "key2": "value2"})
//
// TODO: @zcf redis 的Hash不支持 层级结构
func (c *Client) SetHash(ctx context.Context, key string, value map[string]interface{}) error {
	return c.client.HMSet(key, value).Err()
}

// GetHashField 获取执行hash的指定field数据
// TODO: @zcf 当field对应的数据为非string数据类型时, 需要对该方法进行改造(例如: 返回[]byte); 如果要改造,需要对redis interface同步修改
func (c *Client) GetHashField(ctx context.Context, key, field string) (string, error) {
	return c.client.HGet(key, field).Result()
}

/*********************************** list接口 ****************************************/
// TODO: @zcf 针对list类型的头部（左边）和尾部（右边）的操作暂时不做多余实现.
// 当前默认, list添加数据从尾部（右边）添加, 从头部（左边）开始获取

// PushList 向key对应的list中从尾部(右边)追加数据
func (c *Client) PushList(ctx context.Context, key string, values ...interface{}) error {
	return c.client.RPush(key, values...).Err()
}

// LenList 获取指定列表长度
func (c *Client) LenList(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(key).Result()
}

// PopList 从头部（左边）开始获取并删除
func (c *Client) PopList(ctx context.Context, key string, value interface{}) error {
	return c.client.LPop(key).Scan(value)
}

/*********************************** set接口 ****************************************/
// TODO: @zcf 当前进提供最基本的操作, 暂时不支持集合运算

// AddSet 向key集合中添加成员
func (c *Client) AddSet(ctx context.Context, key string, values ...interface{}) error {
	return c.client.SAdd(key, values...).Err()
}

// CheckSetMember 检查成员是否在集合内
func (c *Client) CheckSetMember(ctx context.Context, key string, value interface{}) (bool, error) {
	return c.client.SIsMember(key, value).Result()
}

// RemSetEle 移除key集合中的元素
func (c *Client) RemSetEle(ctx context.Context, key string, values ...interface{}) error {
	return c.client.SRem(key, values...).Err()
}

/*********************************** 有序集合接口 ****************************************/

// AddZSet 向key有序集合添加成员
// members map[成员数据]分数
func (c *Client) AddZSet(ctx context.Context, key string, members ...*ZSetMember) error {
	zms := make([]*redis.Z, 0, len(members))
	for _, m := range members {
		zms = append(zms, &redis.Z{Score: m.Score, Member: m.Member})
	}
	return c.client.ZAdd(key, zms...).Err()
}

// CardZSet 获取有序集合的成员数
func (c *Client) CardZSet(ctx context.Context, key string) (int64, error) {
	return c.client.ZCard(key).Result()
}

// MembersWithScoreZSet 从高到低获取有序集合的成员及分数
func (c *Client) MembersWithScoreZSet(ctx context.Context, key string) ([]*ZSetMember, error) {
	zres, err := c.client.ZRevRangeWithScores(key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	values := make([]*ZSetMember, 0, len(zres))
	err = copier.Copy(&values, &zres)
	return values, err
}

// RemMembersZSet 移除key有序集合中的members
func (c *Client) RemMembersZSet(ctx context.Context, key string, members ...string) error {
	return c.client.ZRem(key, members).Err()
}

/*********************************** pubsub接口 ****************************************/
func (c *Client) Publish(ctx context.Context, key string, value interface{}) error {
	return c.client.Publish(key, value).Err()
}

func (c *Client) Subscribe(ctx context.Context, channels ...string) (<-chan *Message, error) {
	pubsub := c.client.Subscribe(channels...)
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
