package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	cvt "github.com/8xmx8/easier/pkg/convert"
	"github.com/8xmx8/easier/pkg/utils"

	"github.com/8xmx8/easier/pkg/logger"
	"github.com/IBM/sarama"
)

/*
原则:
Consumer 加入给定Topic列表的 消费者集群, 并通过ConsumerHandler启动, 并阻塞ConsumerGroupSession;
会话的生命周期如下:
1. Consumer加入 Consumer Group, 并分配给他们"合理份额"的分区, 又名:claims;
2. 在开始处理之前, 会调用Handler的setup()钩子来通知用户claims, 并允许对状态进行任何必要的准备或更改;
3. 对于每个分配的claims, Handler的ConsumeClaim()将在一个单独的goroutine中调用, 该goroutine要求它是线程安全的; 必须小心的保护任何状态, 防止并发读/写;
4. 会话将持续, 直到其中一个consumerClaim()退出; 这可以是在"取消父上下文"时，也可以是"在启动服务器端rebalance周期"时;
5. 一旦所有consumerClaim()循环退出, 就会调用处理程序的Cleanup()钩子, 以允许用户在rebalance之前执行任何最终任务;
6. 最后，在claims released之前，最后一次提交标记的offset;

PS: 一旦触发rebalance，会话必须在Config中完成; 这意味着ConsumerClaim()必须尽快退出, 以便有时间进行Cleanup()和最终的offset commit;
如果超过超时, Kafka会将使用者从组中删除, 这将导致偏移提交失败;
*/

// ConsumerGroup 定义消费者组类
type ConsumerGroup struct {
	logger   logger.Logger
	consumer sarama.ConsumerGroup
	conf     *sarama.Config
	options  consumerOption
}

type Data struct {
	FaninTime time.Time
	Topic     string
	Key       []byte
	Value     []byte
}

// String data to string
func (d *Data) String() string {
	return fmt.Sprintf("{ Topic: %s, Key: %s, Value: %s, FaninTime:%s }",
		d.Topic, d.Key, d.Value, d.FaninTime.Format(utils.StrTimeFormatMill))
}

// consumerOption 定义创建消费者所需要的参数
// 并实现了 `ConsumerGroupHandler`
type consumerOption struct {
	logger   logger.Logger
	callback func(context.Context, *Data) error
	groupID  string
	topics   []string
	addrs    []string
}

// nolint
func (co consumerOption) Setup(sess sarama.ConsumerGroupSession) error {
	co.logger.Info("consumer setup", logger.MakeField("groupID", co.groupID), logger.MakeField("{topic: partition}", sess.Claims()))
	return nil
}

// nolint
func (co consumerOption) Cleanup(sess sarama.ConsumerGroupSession) error {
	co.logger.Info("consumer exiting", logger.MakeField("groupID", co.groupID))
	sess.Commit()
	return nil
}

func (co consumerOption) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error { // nolint
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok || msg == nil {
				co.logger.Error(logger.ErrorKafkaConsumer, "消费到空数据", logger.MakeField("groupID", co.groupID), logger.MakeField("data", msg))
				continue
			}
			// Mark: @zcf Callback崩了会影响到消费任务
			// 所以, Callback一定要用好Context
			data := &Data{
				Key:       msg.Key,
				Value:     msg.Value,
				Topic:     msg.Topic,
				FaninTime: msg.Timestamp,
			}
			err := co.callback(sess.Context(), data)
			retry := 0
			for err != nil && retry < 5 {
				co.logger.Error(logger.ErrorKafkaConsumer, "callback handle", logger.MakeField("GroupID", co.groupID),
					logger.MakeField("Topic", claim.Topic()), logger.ErrorField(err))
				if err = co.callback(sess.Context(), data); err == nil {
					break
				}
				<-time.After(time.Second)
				retry++
			}
			co.logger.Infof("consumer[%s]: Topic: %s, Partition: %v, Offset:%v, Accession: %s ago",
				co.groupID, claim.Topic(), claim.Partition(), msg.Offset, time.Since(msg.Timestamp).String())
			sess.MarkMessage(msg, "")
		case <-sess.Context().Done():
			co.logger.Info("Consumer context closing", logger.MakeField("GroupID", co.groupID),
				logger.MakeField("Topic", claim.Topic()), logger.MakeField("Partition", claim.Partition()),
				logger.MakeField("NextOffset", claim.HighWaterMarkOffset()),
			)
			return nil
		}
	}
}

type HandleFunc func(context.Context, *Data) error

// Fetch 定义消费者拉取缓存的配置对象
type Fetch struct {
	Min, Default, Max int32
}

func ConsumerWithFetch(f Fetch) OptionFunc {
	return func(c *Config) error {
		c.conf.Consumer.Fetch = f
		return nil
	}
}

func ConsumerBeginAt(at int64) OptionFunc {
	return func(c *Config) error {
		c.conf.Consumer.Offsets.Initial = at
		return nil
	}
}

// NewConsumerGroup 创建一个消费者实例
// 默认从最旧的开始消费
func NewConsumerGroup(addrs, topics []string, groupID string, handle HandleFunc, ops ...OptionFunc) (*ConsumerGroup, error) {
	conf := DefaultConfig()
	conf.conf.Consumer.Return.Errors = true
	conf.conf.Consumer.Offsets.Initial = sarama.OffsetOldest
	conf.conf.Consumer.Fetch = Fetch{
		Min:     1 << 10,  // 1KB
		Default: 10 << 20, // 10MB
		Max:     20 << 20, // 20 MB
	}
	conf.conf.Consumer.Group.Session.Timeout = 2 * time.Minute
	conf.conf.Consumer.Offsets.AutoCommit.Enable = true              // 自动提交offset(默认开启)
	conf.conf.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second // 默认 1s
	for _, op := range ops {
		if err := op(conf); err != nil {
			return nil, err
		}
	}
	cg, err := sarama.NewConsumerGroup(addrs, groupID, conf.conf)
	if err != nil {
		return nil, err
	}
	consumer := &ConsumerGroup{
		conf:     conf.conf,
		logger:   conf.logger,
		consumer: cg,
		options: consumerOption{
			logger:   conf.logger,
			callback: handle,
			groupID:  groupID,
			topics:   topics,
			addrs:    addrs,
		},
	}
	return consumer, nil
}

// Run 执行消费动作
func (cg *ConsumerGroup) Run(ctx context.Context) error {
	cg.logger.Info("run kafka consumer", logger.MakeField("topics", cg.options.topics),
		logger.MakeField("groupID", cg.options.groupID), logger.MakeField("addrs", cg.options.addrs))
	defer cg.close()
	for {
		select {
		case err := <-cg.consumer.Errors(): // Track errors
			cg.logger.Error(logger.ErrorKafkaConsumer, "Error channel", logger.MakeField("GroupId", cg.options.groupID),
				logger.ErrorField(err))
		case <-ctx.Done():
			err := ctx.Err()
			if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				// 正常退出
				return nil
			}
			cg.logger.Error(logger.ErrorMethod, "上下文异常退出", logger.MakeField("GroupId", cg.options.groupID),
				logger.ErrorField(err))
			return err
		default:
			if err := cg.consumer.Consume(ctx, cg.options.topics, cg.options); err != nil {
				cg.logger.Error(logger.ErrorKafkaConsumer, "Error channel",
					logger.MakeField("GroupId", cg.options.groupID), logger.ErrorField(err))
				// TODO: 这个错误需要关注一下,暂时不知道这里报错会造成什么问题
				return err
			}
		}
	}
}

func (cg *ConsumerGroup) close() {
	if cvt.IsNil(cg.consumer) {
		if err := cg.consumer.Close(); err != nil {
			cg.logger.Error(logger.ErrorKafkaConsumer, "Consumer stoped error",
				logger.MakeField("GroupId", cg.options.groupID), logger.ErrorField(err))
			return
		}
	}
	cg.logger.Info("Consumer stoped", logger.MakeField("GroupId", cg.options.groupID))
}
