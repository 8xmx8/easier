package es

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/8xmx8/easier/pkg/logger"
	"github.com/olivere/elastic/v7"
)

//go:embed mapping
var Mapping embed.FS

func LoadMapping(filename string) map[string]interface{} {
	mapping := map[string]interface{}{}
	bytes, err := Mapping.ReadFile(filename)
	if err != nil {
		return mapping
	}
	err = json.Unmarshal(bytes, &mapping)
	if err != nil {
		return mapping
	}
	return mapping
}

// CreateDoc 创建一个文档
func (e *Elastic) CreateDoc(ctx context.Context, index string, data interface{}) error {
	// if !e.IsExistsIndex(ctx, index) {
	// 	return fmt.Errorf("check index [%s] is not exists", index)
	// }
	put, err := e.es.Index().Index(index).BodyJson(data).Do(ctx)
	if err != nil {
		return err
	}
	e.loggeer.Infof("indexed [%s-%s] to index [%s], type [%s]", index, put.Id, put.Index, put.Type)
	return nil
}

// CreateDocByID 创建一个文档
func (e *Elastic) CreateDocByID(ctx context.Context, index, id string, data interface{}) error {
	// if !e.IsExistsIndex(ctx, index) {
	// 	return fmt.Errorf("check index [%s] is not exists", index)
	// }
	put, err := e.es.Index().Index(index).Id(id).BodyJson(data).Do(ctx)
	if err != nil {
		return err
	}
	e.loggeer.Infof("indexed [%s-%s] to index [%s], type [%s]", index, put.Id, put.Index, put.Type)
	return nil
}

// SyncBatchCreateDoc 同步的批量写入
// TODO: 后期需要限制一下批量数量
func (e *Elastic) SyncBatchCreateDoc(ctx context.Context, index string, datas []interface{}) error {
	bulkRequest := e.es.Bulk()
	for _, item := range datas {
		bulkRequest = bulkRequest.Add(elastic.NewBulkIndexRequest().Type("_doc").Index(index).Doc(item))
	}
	res, err := bulkRequest.Do(ctx)
	if err != nil {
		return err
	}
	if res.Errors {
		e.loggeer.Errorf(logger.ErrorES, "sync batch create doc: %+v", res.Items)
		return errors.New("sync batch create doc error")
	}
	e.loggeer.Info("sync batch create doc success")
	return nil
}

type AsyncHandle elastic.BulkAfterFunc

type asyncBatchCreateOption struct {
	bulkActions   int           // 每个携程队列容量
	flushInterval time.Duration // 刷新间隔
	workerNum     int           // 携程数
	stats         bool          // 是否获取统计信息
}
type AsyncBatchCreateOptionFunc func(abc *asyncBatchCreateOption)

func WithBulkActions(n int) AsyncBatchCreateOptionFunc {
	return func(abc *asyncBatchCreateOption) {
		abc.bulkActions = n
	}
}

func WithFlushInterval(t time.Duration) AsyncBatchCreateOptionFunc {
	return func(abc *asyncBatchCreateOption) {
		abc.flushInterval = t
	}
}
func WithWorkerNum(n int) AsyncBatchCreateOptionFunc {
	return func(abc *asyncBatchCreateOption) {
		abc.workerNum = n
	}
}
func WithStats(n bool) AsyncBatchCreateOptionFunc {
	return func(abc *asyncBatchCreateOption) {
		abc.stats = n
	}
}

type Option int

const (
	Default Option = iota // 0. 空处理
	Insert                // 1. 新增
	Update                // 2. 更新
	Delete                // 3. 删除
)

type DatContainer struct {
	Data  interface{} `json:"data"`
	ID    string      `json:"id"`
	Index string      `json:"index"`
	Opt   Option      `json:"option"`
}

// SyncBatchCreateDoc 异步的批量写入
func (e *Elastic) AsyncBatchCreateDoc(ctx context.Context, datas <-chan *DatContainer,
	handle AsyncHandle, ops ...AsyncBatchCreateOptionFunc) error {
	abc := &asyncBatchCreateOption{
		bulkActions:   100,
		flushInterval: time.Second,
		workerNum:     10,
		stats:         false,
	}
	for _, op := range ops {
		op(abc)
	}
	ep, err := e.es.BulkProcessor().
		BulkActions(abc.bulkActions).
		FlushInterval(abc.flushInterval).     // 刷新间隔
		Workers(abc.workerNum).               // 携程数
		Stats(abc.stats).                     // 是否获取统计信息
		After(elastic.BulkAfterFunc(handle)). // 刷新后回调函数
		Do(ctx)
	if err != nil {
		return err
	}
	if err := ep.Start(ctx); err != nil {
		return err
	}
	defer ep.Close() // 关闭并提交所有队列里的数据，一定要做
	for {
		select {
		case data, ok := <-datas:
			if !ok {
				continue
			}
			switch data.Opt {
			case Insert:
				query := elastic.NewBulkIndexRequest().Type("_doc").Index(data.Index).Doc(data.Data)
				if data.ID != "" {
					query.Id(data.ID)
				}
				ep.Add(query)
				elastic.NewBulkCreateRequest()
			case Update:
				if data.ID == "" {
					e.loggeer.Info("[delete] no ID", logger.MakeField("index", data.Index), logger.MakeField("new doc", data.Index))
					continue
				}
				// 使用Doc方法进行部分更新
				// data.Data 应该是一个map
				query := elastic.NewBulkUpdateRequest().Index(data.Index).Id(data.ID).Doc(data.Data)
				ep.Add(query)
			case Delete:
				if data.ID == "" {
					e.loggeer.Info("[delete] no ID", logger.MakeField("index", data.Index))
					continue
				}
				query := elastic.NewBulkDeleteRequest().Index(data.Index).Id(data.ID)
				ep.Add(query)
			default:
				query := elastic.NewBulkIndexRequest().Type("_doc").Index(data.Index).Doc(data.Data)
				if data.ID != "" {
					query.Id(data.ID)
				}
				ep.Add(query)
			}
		case <-ctx.Done():
			e.loggeer.Info("async batch create doc exit")
			return nil
		}
	}
}

// DocIsExist 检查文档是否存在
// https://www.elastic.co/guide/cn/elasticsearch/guide/current/doc-exists.html
func (e *Elastic) DocIsExist(ctx context.Context, index, id string) (isExist bool) {
	path := fmt.Sprintf("/%s/_doc/%s", index, id) // "/{index}/{type}/{id}""
	resp, err := e.es.PerformRequest(ctx, elastic.PerformRequestOptions{
		Method:       http.MethodHead,
		Path:         path,
		IgnoreErrors: []int{},
	})
	if err != nil && !elastic.IsNotFound(err) {
		e.loggeer.Error(logger.ErrorES, "检查文档是否存在", logger.ErrorField(err), logger.MakeField(index, id))
		return
	}
	return (resp.StatusCode == http.StatusOK)
}

// DocIsExistByQuery 检查文档是否存在
func (e *Elastic) DocIsExistByQuery(ctx context.Context, index string, query elastic.Query) (isExist bool) {
	searchService := e.es.Search().Index(index)
	resp, err := searchService.Query(query).Do(ctx)
	if err != nil {
		return false
	}
	return resp.Hits.TotalHits.Value > 0
}
