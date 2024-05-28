package cache

import (
	"context"
	"errors"
	"time"
)

// 声明对list的操作点
type ListFlag int8

const (
	StartPoint ListFlag = iota + 1 // 列表的头部（左边）
	EndPoint                       // 列表的尾部（右边）
)

// 声明支持的redis部署模式
type DeployMode string

const (
	ClientMod   DeployMode = "client"   // 单机模式
	ClusterMod  DeployMode = "cluster"  // 集群模式
	SentinelMod DeployMode = "sentinel" // 哨兵模式
)

// ZSetMember 有序集合的数据结构
type ZSetMember struct { // nolint
	Score  float64     `json:"score"`
	Member interface{} `json:"member"`
}

type Message struct {
	Channel      string
	Pattern      string
	Payload      interface{}
	PayloadSlice []string
}

type Redis interface {
	IsExist(ctx context.Context, key ...string) bool
	FlushDB(ctx context.Context, isAll bool) error
	Del(ctx context.Context, key ...string) error
	SetExpire(ctx context.Context, key string, ttl time.Duration) error
	GetExpire(ctx context.Context, key string) (time.Duration, error)
	GetMixed(ctx context.Context, key string, value interface{}) error
	ScanKey(ctx context.Context, match string) chan string

	SetStr(ctx context.Context, key, value string) error
	SetStrTTL(ctx context.Context, key, value string, ttl time.Duration) error
	SetNX(ctx context.Context, key, value string, ttl time.Duration) error

	SetHash(ctx context.Context, key string, value map[string]interface{}) error
	GetHashField(ctx context.Context, key, field string) (string, error)

	PushList(ctx context.Context, key string, values ...interface{}) error
	LenList(ctx context.Context, key string) (int64, error)
	PopList(ctx context.Context, key string, value interface{}) error

	AddSet(ctx context.Context, key string, values ...interface{}) error
	CheckSetMember(ctx context.Context, key string, value interface{}) (bool, error)
	RemSetEle(ctx context.Context, key string, values ...interface{}) error

	AddZSet(ctx context.Context, key string, members ...*ZSetMember) error
	CardZSet(ctx context.Context, key string) (int64, error)
	MembersWithScoreZSet(ctx context.Context, key string) ([]*ZSetMember, error)
	RemMembersZSet(ctx context.Context, key string, members ...string) error
	Publish(ctx context.Context, channel string, message interface{}) error
	Subscribe(ctx context.Context, channels ...string) (<-chan *Message, error)
}

// RedisConf redis的连接配置
// nolint
type RedisConf struct {
	DeployMode DeployMode `json:"deployMode"`
	Endpoints  []string   `json:"endpoints"`
	User       string     `json:"user"`
	Password   string     `json:"passwd"`
	Db         int        `json:"db"`
}

// InitRedisClient 实例化redis连接对象
func InitRedisClient(ctx context.Context, conf *RedisConf) (client Redis, err error) {
	switch conf.DeployMode {
	case SentinelMod: // 哨兵模式
		return nil, errors.New("不支持'哨兵模式'")
	case ClusterMod: // 集群模式
		if client, err = NewCluster(ctx, conf.Endpoints,
			ClusterWithAuth(conf.User, conf.Password),
		); err != nil {
			return
		}
	default: // 默认单机模式
		if client, err = NewClient(ctx, conf.Endpoints[0], conf.Db,
			WithAuth(conf.User, conf.Password),
		); err != nil {
			return
		}
	}
	return
}
