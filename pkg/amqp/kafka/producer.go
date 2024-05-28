package kafka

import (
	"context"
	"sync"
	"time"

	cvt "github.com/8xmx8/easier/pkg/convert"

	"github.com/8xmx8/easier/pkg/logger"
	"github.com/IBM/sarama"
)

const (
	maxMsgBytes = 10 << 20 // 10MB
)

type Producer struct {
	conf     *sarama.Config
	producer sarama.SyncProducer // 阻塞的生产者
	logger   logger.Logger
	msgBatch int
}

func ProducerWithBatchSize(size int) OptionFunc {
	return func(c *Config) error {
		c.producerMsgBatch = size
		return nil
	}
}
func ProducerWithMaxMessageBytes(size int) OptionFunc {
	return func(c *Config) error {
		c.conf.Producer.MaxMessageBytes = size
		return nil
	}
}

func NewProducer(addrs []string, ops ...OptionFunc) (*Producer, error) {
	conf := DefaultConfig()
	conf.conf.Version = defaultVersion
	// 采用随机而非哈希方法
	// https://pkg.go.dev/github.com/Shopify/sarama@v1.32.0#Partitioner
	conf.conf.Producer.Partitioner = sarama.NewRandomPartitioner
	conf.conf.Producer.Return.Successes = true
	conf.conf.Producer.Return.Errors = true
	conf.conf.Producer.Idempotent = true // 开启幂等性
	conf.conf.Net.MaxOpenRequests = 1    // 开启幂等性后 并发请求数也只能为1
	conf.conf.Producer.RequiredAcks = sarama.WaitForAll
	conf.conf.Producer.Timeout = time.Duration(5) * time.Minute
	conf.conf.Producer.MaxMessageBytes = maxMsgBytes
	for _, op := range ops {
		if err := op(conf); err != nil {
			return nil, err
		}
	}

	p, err := sarama.NewSyncProducer(addrs, conf.conf)
	if err != nil {
		return nil, err
	}
	prod := &Producer{
		conf:     conf.conf,
		logger:   conf.logger,
		producer: p,
		msgBatch: conf.producerMsgBatch,
	}
	return prod, nil
}

// AsyncPusher 使用通道向kafka发送数据
// TODO: @zcf discuss:这里提供了基于channel的异步生产方式, 但是生产时会产生错误, 这里错误直接打印了;
// 如果后期有需要, 是否可以考虑加一个用来接收error的channel
func (p *Producer) AsyncPusher(ctx context.Context, msgChan <-chan *Msg) {
	msgList := &msgList{msgs: make([]*sarama.ProducerMessage, 0, p.msgBatch)}
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		select {
		case msg := <-msgChan:
			// if msg == nil || msg.Key == "" || len(msg.Value) == 0 {
			if msg == nil || len(msg.Value) == 0 {
				continue
			}
			msgList.push(msg.makeProducMsg())
			if msgList.count() >= p.msgBatch {
				if err := p.sendBatch(msgList.getMsgs()); err != nil {
					p.logger.Error(logger.ErrorKafkaProducerSend, "kafka producer send data", logger.ErrorField(err))
				}
				msgList.clear()
			}
		case <-tick.C:
			if err := p.sendBatch(msgList.getMsgs()); err != nil {
				p.logger.Error(logger.ErrorKafkaProducerSend, "kafka producer send data for ticker", logger.ErrorField(err))
			}
			msgList.clear()
		case <-ctx.Done(): // 上下文停止
			p.logger.Info("producer is stopping")
			return
		}
	}
}

// PushMsgs 批量的向kafka的同一个topic发送数据
// PS: @zcf这里不对推送数量数量做限制和校验
func (p *Producer) BatchMsgsPush(ctx context.Context, msgs []*Msg) error {
	msgPack := &msgList{msgs: make([]*sarama.ProducerMessage, 0, p.msgBatch)}
	for _, msg := range msgs {
		if msg == nil || msg.Key == "" || len(msg.Value) == 0 {
			p.logger.Error(logger.ErrorParamsIncomplete, "msg is nil", logger.MakeField("data", msg))
			continue
		}
		msgPack.push(msg.makeProducMsg())
	}
	return p.sendBatch(msgPack.getMsgs())
}

// SinglePushMsg 向kafka的一个topic发送一个数据
func (p *Producer) SingleMsgPush(ctx context.Context, msg *Msg) (err error) {
	if msg == nil || msg.Key == "" || len(msg.Value) == 0 {
		p.logger.Error(logger.ErrorParamsIncomplete, "msg is nil", logger.MakeField("data", msg))
		return
	}
	_, _, err = p.producer.SendMessage(msg.makeProducMsg())
	return
}

func (p *Producer) sendBatch(msgList []*sarama.ProducerMessage) error {
	if len(msgList) < 1 {
		return nil
	}
	if err := p.producer.SendMessages(msgList); err != nil {
		errs, ok := err.(sarama.ProducerErrors)
		if ok {
			resErrs := make(sarama.ProducerErrors, 0)
			for _, item := range errs {
				if _, _, itemErr := p.producer.SendMessage(item.Msg); itemErr != nil {
					p.logger.Infof("singleMsg send error => topic[%s]; error: %s",
						item.Msg.Topic, itemErr)
					item.Err = itemErr
					resErrs = append(resErrs, item)
				}
			}
			return resErrs
		}
		return err
	}
	return nil
}

func (p *Producer) Close() {
	if cvt.IsNil(p.producer) {
		if err := p.producer.Close(); err != nil {
			p.logger.Error(logger.ErrorKafkaProducer, "producer close", logger.ErrorField(err))
		}
	}
}

// Msg 定义消息对象
type Msg struct {
	Topic string
	Key   string
	Value []byte
}

// makeProducMsg 构建ProducerMessage
func (m *Msg) makeProducMsg() *sarama.ProducerMessage {
	return &sarama.ProducerMessage{
		Topic: m.Topic,
		Key:   sarama.StringEncoder(m.Key),
		Value: sarama.ByteEncoder(m.Value),
	}
}

// msgList 定义待发送数据的对象
type msgList struct {
	msgs []*sarama.ProducerMessage
	lock sync.Mutex
}

func (s *msgList) push(msg *sarama.ProducerMessage) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.msgs = append(s.msgs, msg)
}

func (s *msgList) count() int {
	return len(s.msgs)
}

func (s *msgList) getMsgs() []*sarama.ProducerMessage {
	return s.msgs
}

func (s *msgList) clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.msgs = []*sarama.ProducerMessage{}
}
