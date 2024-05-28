package kafka

import (
	"context"
	"fmt"

	"github.com/8xmx8/easier/pkg/logger"
	"github.com/8xmx8/easier/pkg/utils"
	"github.com/IBM/sarama"
)

var (
	defaultVersion = sarama.V0_11_0_0
)

var (
	topicTTLPolicy         = "delete"
	topicTTLRetention      = "259200000" // 3d->259200000ms; 2d ->172800000ms;
	topicNumPartitions     = int32(10)   //
	topicReplicationFactor = int16(1)    // 副本数
)

type OptionFunc func(*Config) error

type Config struct {
	conf             *sarama.Config
	logger           logger.Logger
	producerMsgBatch int // 生产者缓冲队列长度
}

func DefaultConfig() *Config {
	conf := &Config{
		conf:             sarama.NewConfig(),
		logger:           logger.DefaultLogger(),
		producerMsgBatch: 100,
	}
	conf.conf.Version = defaultVersion
	return conf
}

func WithVersion(v sarama.KafkaVersion) OptionFunc {
	return func(c *Config) error {
		c.conf.Version = v
		return nil
	}
}

func WithLogger(log logger.Logger) OptionFunc {
	return func(c *Config) error {
		c.logger = log
		return nil
	}
}

// Client 定义KafkaClient
type Client struct {
	conf *Config
	logg logger.Logger
	cli  sarama.ClusterAdmin
}
type ClientOptionFunc func(*Client) error

func ClientWithLogger(logg logger.Logger) ClientOptionFunc {
	return func(c *Client) error {
		c.logg = logg
		return nil
	}
}

// NewKafkaClient 实例化kafka连接实例
// retrun: kafkaClient, kafkaCloseFunc, error
func NewKafkaClient(addrs []string, ops ...OptionFunc) (*Client, func(), error) {
	conf := DefaultConfig()
	for _, op := range ops {
		if err := op(conf); err != nil {
			return nil, nil, err
		}
	}
	cli, err := sarama.NewClusterAdmin(addrs, conf.conf)
	if err != nil {
		return nil, func() {}, err
	}
	kfk := &Client{
		conf: conf,
		logg: conf.logger,
		cli:  cli,
	}
	return kfk, kfk.close, nil
}

func (kc *Client) close() {
	if kc.cli != nil {
		_ = kc.cli.Close()
	}
}

// GetTopics 获取当前kafka中的topic
func (kc *Client) GetTopics(ctx context.Context) (topicList []string, err error) {
	curTopics, err := kc.cli.ListTopics()
	if err != nil || len(curTopics) == 0 {
		return
	}
	for name := range curTopics {
		topicList = append(topicList, name)
	}
	return
}

// GetBrokers 获取集群中的borker list
func (kc *Client) GetBrokers(ctx context.Context) (brokers []*sarama.Broker, err error) {
	brokers, _, err = kc.cli.DescribeCluster()
	return
}

// //////////////////////////////////////////////////////////////////////////////////////////
// TopicOption 创建topic的配置
type TopicOption func(*sarama.TopicDetail)

func TopicWithOfPartitions(num int32) TopicOption {
	return func(topic *sarama.TopicDetail) {
		topic.NumPartitions = num
	}
}

func TopicWithNumOfReplication(num int16) TopicOption {
	return func(topic *sarama.TopicDetail) {
		topic.ReplicationFactor = num
	}
}
func TopicWithConfigEntries(key, value string) TopicOption {
	return func(topic *sarama.TopicDetail) {
		topic.ConfigEntries[key] = &value
	}
}

type CreateTopicsErr struct {
	errs map[string]error
}

func NewCreateTopicsErr(errs map[string]error) *CreateTopicsErr {
	return &CreateTopicsErr{errs: errs}
}
func (err CreateTopicsErr) Error() string {
	return fmt.Sprintf("%+v", err.errs)
}

// CreateTopics 创建topics
func (kc *Client) CreateTopics(ctx context.Context, tarTopics []string, ops ...TopicOption) (err error) {
	tops, err := kc.GetTopics(ctx)
	if err != nil {
		return NewCreateTopicsErr(map[string]error{"all": err})
	}
	topSet := utils.NewStringSet(tops...)
	errs := map[string]error{}
	for _, name := range tarTopics {
		if topSet.Has(name) {
			kc.logg.Info("topic已存在, 无法重复创建", logger.MakeField("topic", name))
			continue
		}
		if err := kc.CreateTopic(ctx, name, ops...); err != nil {
			errs[name] = err
		}
		kc.logg.Info("topic创建成功", logger.MakeField("topic", name))
	}
	if len(errs) > 0 {
		return NewCreateTopicsErr(errs)
	}
	return
}

// CreateTopic 创建单个Topic
func (kc *Client) CreateTopic(ctx context.Context, topicName string, ops ...TopicOption) (err error) {
	topicConf := &sarama.TopicDetail{
		NumPartitions:     topicNumPartitions,
		ReplicationFactor: topicReplicationFactor,
		ConfigEntries: map[string]*string{
			"cleanup.policy": &topicTTLPolicy,
			"retention.ms":   &topicTTLRetention,
		},
	}
	for _, op := range ops {
		op(topicConf)
	}
	return kc.cli.CreateTopic(topicName, topicConf, false)
}
