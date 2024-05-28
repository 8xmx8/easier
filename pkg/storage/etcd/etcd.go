package etcd

import (
	"context"
	"errors"
	"time"

	"github.com/8xmx8/easier/pkg/utils"
	etcdcli "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.uber.org/zap"
)

const (
	optTimeout    = 5 * time.Second
	dialTimeout   = 5 * time.Second
	dialAliveTime = 5 * time.Second
)

var (
	ErrInvalidParam = errors.New("invalid param")
	nodeName        = "etcd-main" // ETCD节点名称
)

type Client struct { // nolint
	nodeName  string
	cli       *etcdcli.Client
	logger    *zap.Logger
	timeout   time.Duration                 // 操作超时
	elections []*Election                   // 选举器
	watchers  []*Watcher                    // 监听器
	locks     map[string]*concurrency.Mutex // 分布式锁
}

func (cli *Client) WithNodeName(name string) *Client {
	cli.nodeName = name
	return cli
}

type OptionFunc func(*etcdcli.Config) error

// WithBaseAuth 基础鉴权
func WithBaseAuth(user, passwd string) OptionFunc {
	return func(cf *etcdcli.Config) error {
		cf.Username = user
		cf.Password = passwd
		return nil
	}
}

// WithDialKeepAliveTime 客户端超时时间
func WithDialKeepAliveTime(time time.Duration) OptionFunc {
	return func(cf *etcdcli.Config) error {
		cf.DialKeepAliveTime = time
		return nil
	}
}

// Withlogger 配置logger
func Withlogger(lg *zap.Logger) OptionFunc {
	return func(cf *etcdcli.Config) error {
		cf.Logger = lg
		return nil
	}
}

// NewETCD 创建ETCD客户端
func NewETCD(ctx context.Context, endpoints []string, ops ...OptionFunc) (*Client, error) {
	log, _ := zap.NewProduction()
	c := &etcdcli.Config{
		Endpoints:         endpoints,
		DialTimeout:       dialTimeout,
		DialKeepAliveTime: dialAliveTime,
		Context:           ctx,
		Logger:            log,
	}
	for _, op := range ops {
		if err := op(c); err != nil {
			return nil, err
		}
	}

	cli, err := etcdcli.New(*c)
	if err != nil {
		return nil, err
	}
	ec := &Client{
		nodeName:  nodeName,
		cli:       cli,
		timeout:   optTimeout,
		logger:    c.Logger,
		elections: []*Election{},
		watchers:  []*Watcher{},
		locks:     map[string]*concurrency.Mutex{},
	}
	return ec, nil
}

// Close 关闭etcd客户端
func (cli *Client) Close() error {
	return cli.cli.Close()
}

// Put向ETCD永久写入一个Key.
func (cli *Client) Put(ctx context.Context, key, value string) (err error) {
	// Get the value first, must be return timeout seconds
	timeoutCtx, cancel := context.WithTimeout(ctx, cli.timeout)
	defer cancel()
	retry := 3
	for {
		_, err = cli.cli.Put(timeoutCtx, key, value)
		if err == nil || retry < 1 {
			break
		}
		cli.logger.Error("数据写入错误", zap.Error(err), zap.Int("retry", retry))
		retry--
		<-time.After(time.Second)
	}
	return
}

// PutWithTTL 向ETCD写入一个带TTL的Key.
func (cli *Client) PutWithTTL(ctx context.Context, key, value string, ttl int64) (err error) {
	retry := 3
	var resp *etcdcli.LeaseGrantResponse
	for {
		resp, err = cli.cli.Grant(ctx, ttl)
		if err == nil || retry < 1 {
			break
		}
		cli.logger.Error("creates a new lease", zap.Error(err), zap.Int("retry", retry))
		retry--
		<-time.After(time.Second)
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, cli.timeout)
	defer cancel()
	retry = 3
	for {
		_, err = cli.cli.Put(timeoutCtx, key, value, etcdcli.WithLease(resp.ID))
		if err == nil || retry < 1 {
			break
		}
		cli.logger.Error("数据协议写入错误", zap.Error(err), zap.Int("retry", retry))
		retry--
		<-time.After(time.Second)
	}

	return err
}

// Get 从ETCD中获取一个KEY，通过前缀获得，所以有可能有多个值
func (cli *Client) Get(ctx context.Context, prefix string) ([]string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cli.timeout)
	resp, err := cli.cli.Get(timeoutCtx, prefix, etcdcli.WithPrefix())
	cancel()
	if err != nil {
		return nil, err
	}
	values := make([]string, 0, resp.Count)
	for _, ev := range resp.Kvs {
		values = append(values, string(ev.Value))
	}
	return values, nil
}

// Pop 从ETCD中弹出一个key
func (cli *Client) Pop(ctx context.Context, prefix string, limit int64) ([]string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cli.timeout)
	resp, err := cli.cli.Get(timeoutCtx, prefix, etcdcli.WithPrefix(), etcdcli.WithLimit(limit))
	cancel()
	if err != nil {
		return nil, err
	}
	values := make([]string, 0, len(resp.Kvs))
	for _, ev := range resp.Kvs {
		if err := cli.Delete(ctx, string(ev.Key)); err != nil {
			cli.logger.Sugar().Warnf("delete key error: %v", err)
		}
		values = append(values, string(ev.Value))
	}
	return values, nil
}

// GetLimit 从ETCD中通过前缀获得，所以有可能有多个值
func (cli *Client) GetLimit(ctx context.Context, prefix string, offset int64) ([]string, []string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cli.timeout)
	resp, err := cli.cli.Get(timeoutCtx, prefix, etcdcli.WithPrefix(), etcdcli.WithLimit(offset))
	cancel()
	if err != nil {
		return nil, nil, err
	}

	keys := make([]string, 0, len(resp.Kvs))
	values := make([]string, 0, len(resp.Kvs))
	for _, ev := range resp.Kvs {
		keys = append(keys, string(ev.Key))
		values = append(values, string(ev.Value))
	}
	return keys, values, nil
}

// GetKvs 从ETCD中获取一个Key前缀的Key:value对应的map集合
func (cli *Client) GetKvs(ctx context.Context, prefix string) (map[string]string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cli.timeout)
	resp, err := cli.cli.Get(timeoutCtx, prefix, etcdcli.WithPrefix())
	cancel()
	if err != nil {
		return nil, err
	}

	kvMap := make(map[string]string, 1)
	for _, ev := range resp.Kvs {
		kvMap[string(ev.Key)] = string(ev.Value)
	}
	return kvMap, nil
}

func (cli *Client) GetKeys(ctx context.Context, prefix string) ([]string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cli.timeout)
	resp, err := cli.cli.Get(timeoutCtx, prefix, etcdcli.WithPrefix())
	cancel()
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(resp.Kvs))
	for _, ev := range resp.Kvs {
		keys = append(keys, string(ev.Key))
	}
	return keys, nil
}

// GetOne 仅获取一个值
func (cli *Client) GetOne(ctx context.Context, prefix string) (string, error) {
	arr, err := cli.Get(ctx, prefix)
	if err != nil {
		return "", err
	}
	if len(arr) > 0 {
		return arr[0], nil
	}
	return "", nil
}

// Delete 删除一个Key
func (cli *Client) Delete(ctx context.Context, key string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, cli.timeout)
	_, err := cli.cli.Delete(timeoutCtx, key)
	cancel()
	return err
}

// DeleteDeleteWithPrefix 使用前缀删除一批Key
func (cli *Client) DeleteWithPrefix(ctx context.Context, key string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, cli.timeout)
	_, err := cli.cli.Delete(timeoutCtx, key, etcdcli.WithPrefix())
	cancel()
	return err
}

// CreateLeaseID 创建租期ID
func (cli *Client) CreateLeaseID(ctx context.Context, ttl int64) (int64, error) {
	resp, err := cli.cli.Grant(ctx, ttl)
	if err != nil {
		return 0, err
	}
	return int64(resp.ID), nil
}

// PutWithLease 使用租期ID创建数据
func (cli *Client) PutWithLease(ctx context.Context, key, value string, lease int64) error {
	_, err := cli.cli.Put(ctx, key, value, etcdcli.WithLease(etcdcli.LeaseID(lease)))
	return err
}

// KeepAliveOnce 给租约进行保活
func (cli *Client) KeepAliveOnce(ctx context.Context, leaseID int64) error {
	if leaseID == 0 {
		return errors.New("keepAliveOnce have to appoint leaseID")
	}
	resp, err := cli.cli.KeepAliveOnce(ctx, etcdcli.LeaseID(leaseID))
	if err != nil {
		return err
	}
	cli.logger.Info("keepAliveOnce", zap.Int64("leaseID", int64(resp.ID)), zap.Int64("TTL", resp.TTL))
	return nil
}

// GetTTLWithLease 获取租约TTL
func (cli *Client) GetTTLWithLease(ctx context.Context, leaseID int64) (int64, error) {
	resp, err := cli.cli.TimeToLive(ctx, etcdcli.LeaseID(leaseID))
	if err != nil {
		return 0, err
	}
	return resp.TTL, nil
}

type Member struct {
	// name is the human-readable name of the member. If the member is not started, the name will be an empty string.
	Name string `json:"name,omitempty"`
	// peerURLs is the list of URLs the member exposes to the cluster for communication.
	PeerURLs   []string `json:"peerURLs,omitempty"`
	ClientURLs []string `json:"clientURLs,omitempty"`
	// ID is the member ID for this member.
	ID uint64 `json:"ID,omitempty"`
}

// ListMembers 节点成员
func (cli *Client) ListMembers(ctx context.Context) ([]*Member, error) {
	tctx, cancel := context.WithTimeout(ctx, cli.timeout)
	defer cancel()
	resp, err := cli.cli.MemberList(tctx)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	r := make([]*Member, 0, len(resp.Members))
	for _, v := range resp.Members {
		r = append(r, &Member{
			ID:         v.GetID(),
			Name:       v.GetName(),
			PeerURLs:   v.GetPeerURLs(),
			ClientURLs: v.GetClientURLs(),
		})
	}
	return r, nil
}

// AddMember 增加节点.例如: peerURLs=[]string{"http://127.0.0.1:2180"}
func (cli *Client) AddMember(ctx context.Context, peerURLs []string) error {
	resp, err := cli.cli.MemberAdd(ctx, peerURLs)
	if err != nil {
		return err
	}
	cli.logger.Info("AddMember", zap.Any("Member", *resp.Member))
	return nil
}

// DelMember 删除节点.
func (cli *Client) DelMember(ctx context.Context, memberID uint64) error {
	resp, err := cli.cli.MemberRemove(ctx, memberID)
	if err != nil {
		return err
	}
	if len(resp.Members) < 1 {
		cli.logger.Warn("DelMember is empty")
		return nil
	}
	cli.logger.Info("DelMember", zap.Any("Member", *resp.Members[0]))
	return nil
}

// //////////////////////////////////////////////////////////////////////////////////////////////////////////

// NewElection 通过ETCD Client创建一个Election.
func (cli *Client) NewElection(ctx context.Context, key string, conf *ElectionConfig, f ElectionFunc) error {
	// concurrency.WithTTL 设置过期时间，可以在leader掉线的时候，
	// 其他节点能快速获取到相关情况，进行再次选举
	cs, err := concurrency.NewSession(cli.cli, concurrency.WithTTL(conf.Timeout))
	if err != nil {
		return err
	}
	el := concurrency.NewElection(cs, key)
	ectx, cancel := context.WithCancel(ctx)
	eleObj := &Election{
		nodeName: cli.nodeName,
		key:      key,
		conf:     conf,
		cs:       cs,
		ele:      el,
		callback: f,
		logger:   cli.logger,
		cancel:   cancel,
	}
	// 交给client持有
	cli.elections = append(cli.elections, eleObj)
	return eleObj.Start(ectx)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////////////

// NewWatcher 通过ETCD Client创建一个Watcher对象，需要输入监听的前缀.
// 异步的
func (cli *Client) NewWatcher(ctx context.Context, prefix string, f WatcherFunc) error {
	if f == nil {
		return ErrInvalidParam
	}
	cctx, cancel := context.WithCancel(ctx)
	pools, err := utils.NewPool(2000)
	if err != nil {
		cancel()
		return err
	}
	w := &Watcher{
		client:   cli,
		prefix:   prefix,
		callback: f,
		cancel:   cancel,
		ctx:      cctx,
		Logger:   cli.logger,
		pools:    pools,
	}
	cli.watchers = append(cli.watchers, w)
	return w.Start(cctx)
}

func (cli *Client) Compact(ctx context.Context, threshold int64, isdefrag bool) {
	// Compact
	compactOpt := etcdcli.WithCompactPhysical()
	m := cli.cli.Maintenance
	endpoints := cli.cli.Endpoints()
	cli.logger.Info("待压缩的ETCD集群节点", zap.Any("endpoints", endpoints))
	for _, host := range endpoints {
		status, err := m.Status(ctx, host)
		if err != nil {
			cli.logger.Error("[etcd]get endpoint status", zap.Error(err), zap.String("endpoint", host))
			continue
		}
		dbSize := status.DbSize
		cli.logger.Info("ETCD集群节点当前数据量", zap.String("endpoint", host), zap.Int("DBsize", int(dbSize/utils.MB)))
		if dbSize < threshold {
			continue
		}
		hashKVResp, err := m.HashKV(ctx, host, 0)
		if err != nil {
			cli.logger.Error("[etcd]get HashKV", zap.Error(err), zap.String("endpoint", host))
		} else {
			rev := hashKVResp.CompactRevision
			_, err = cli.cli.Compact(ctx, rev, compactOpt)
			if err != nil {
				cli.logger.Error("[etcd]get Compact", zap.Error(err), zap.String("endpoint", host), zap.Int64("CompactRevision", rev))
			} else {
				cli.logger.Info("[etcd]Compact success", zap.String("endpoint", host), zap.Int64("CompactRevision", rev))
			}
		}

		if !isdefrag {
			continue
		}
		// defrag
		_, err = m.Defragment(ctx, host)
		if err != nil {
			cli.logger.Error("[etcd]defragment", zap.Error(err), zap.String("endpoint", host))
			continue
		} else {
			cli.logger.Info("[etcd]defrag success", zap.String("endpoint", host))
		}

		// alarm
		// TODO: 暂时没有需求做"报警"删除
	}
}
