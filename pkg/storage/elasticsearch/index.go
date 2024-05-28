package es

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/8xmx8/easier/pkg/logger"
)

const (
	shards           = "6" // 分片
	replicas         = "0" // 默认单节点的没有副本
	maxWindow        = "2000000"
	totalFieldsLimit = "10000"
)

var (
	ESDateFormat   = "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd||epoch_millis" // 创建index定义date类型时的格式
	TimeDateFormat = "2006-01-02 15:04:05"                           // 存储到es date类型时需要的格式
)

type totalFields struct {
	Limit string `json:"limit"`
}

type settingMapping struct {
	TotalFields *totalFields `json:"total_fields"`
}

type indexSetting struct { // nolint
	Shards    string                 `json:"number_of_shards"`
	Replicas  string                 `json:"number_of_replicas"`
	MaxWindow string                 `json:"max_result_window"`
	Analysis  map[string]interface{} `json:"analysis,omitempty"`
	Mapping   *settingMapping        `json:"mapping"`
}

type IndexOption func(*indexStruct)

// WithShards 配置分片
func WithShards(s int) IndexOption {
	return func(is *indexStruct) {
		is.Settings.Shards = strconv.Itoa(s)
	}
}

// WithReplicas 配置副本数
func WithReplicas(r int) IndexOption {
	return func(is *indexStruct) {
		is.Settings.Replicas = strconv.Itoa(r)
	}
}

// WithMaxWindow 配置最大结果条数
func WithMaxWindow(m int) IndexOption {
	return func(is *indexStruct) {
		is.Settings.MaxWindow = strconv.Itoa(m)
	}
}

// WithLimitFields 配置索引字段最大条数
func WithLimitFields(l int) IndexOption {
	return func(is *indexStruct) {
		is.Settings.Mapping.TotalFields.Limit = strconv.Itoa(l)
	}
}

func WithAliases(as []string) IndexOption {
	return func(is *indexStruct) {
		for _, a := range as {
			is.Aliases[a] = map[int]int{}
		}
	}
}

// WithAnalysis 配置解析器
func WithAnalysis(analysis map[string]interface{}) IndexOption {
	return func(is *indexStruct) {
		is.Settings.Analysis = analysis
	}
}

type indexStruct struct { // nolint
	Settings *indexSetting          `json:"settings"`
	Aliases  map[string]interface{} `json:"aliases,omitempty"`
	Mappings interface{}            `json:"mappings"`
}

// CreateIndex 创建索引
func (e *Elastic) CreateIndex(ctx context.Context, name string, mapping interface{}, ops ...IndexOption) error {
	indexStruct := &indexStruct{
		Settings: &indexSetting{
			Shards:    shards,
			Replicas:  replicas, // 默认单节点的没有副本
			MaxWindow: maxWindow,
			Mapping:   &settingMapping{&totalFields{Limit: totalFieldsLimit}},
		},
		// Mappings: map[string]interface{}{"_doc": mapping}, // MARK: 向下兼容低版本
		Mappings: mapping, // MARK: 7.15.x
	}
	for _, op := range ops {
		op(indexStruct)
	}
	if e.IsExistsIndex(ctx, name) {
		return fmt.Errorf("[%s] index already exists", name)
	}
	createIndex, err := e.es.CreateIndex(name).BodyJson(indexStruct).Do(ctx)
	if err != nil {
		return err
	}
	if !createIndex.Acknowledged {
		return fmt.Errorf("[%s] index creation failed", name)
	}
	return nil
}

func (e *Elastic) IsExistsIndex(ctx context.Context, index string) bool {
	exists, err := e.es.IndexExists(index).Do(ctx)
	if err != nil {
		e.loggeer.Error(logger.ErrorES, "check index is exists error", logger.MakeField("index", index), logger.ErrorField(err))
		return false
	}
	return exists
}

// DeleteIndex 删除索引
func (e *Elastic) DeleteIndex(ctx context.Context, index string) error {
	res, err := e.es.DeleteIndex(index).Do(ctx)
	if !res.Acknowledged {
		return fmt.Errorf("[%s] index delete failed", index)
	}
	return err
}

// UpdateMappingByReIndex 通过 ReIndex 方式修改 mapping
func (e *Elastic) UpdateMappingByReIndex(ctx context.Context, name string, mapping interface{}) error {
	// 1. 创建一个和原始索引一样的临时索引
	now := time.Now().Unix()
	tmpIndex := name + "." + strconv.FormatInt(now, 10)
	oldMapping, err := e.GetMapping(ctx, name)
	if err != nil {
		return err
	}
	err = e.CreateIndex(ctx, tmpIndex, oldMapping)
	if err != nil {
		return err
	}
	// 2. 将数据迁移到临时索引
	_, err = e.es.Reindex().SourceIndex(name).DestinationIndex(tmpIndex).Do(ctx)
	if err != nil {
		return err
	}
	// 3. 删除原始索引
	err = e.DeleteIndex(ctx, name)
	if err != nil {
		return err
	}
	// 4. 创建新索引
	err = e.CreateIndex(ctx, name, mapping)
	if err != nil {
		return err
	}
	// 5. 将临时索引数据迁移回新索引
	err = e.ReIndex(ctx, tmpIndex, name)
	if err != nil {
		return err
	}
	// 6. 将临时索引删除
	err = e.DeleteIndex(ctx, tmpIndex)
	if err != nil {
		return err
	}
	return nil
}

// UpdateMapping 修改 Mapping
func (e *Elastic) UpdateMapping(ctx context.Context, name string, mapping map[string]interface{}) error {
	resp, err := e.es.PutMapping().Index(name).BodyJson(mapping).Do(ctx)
	if err != nil {
		return err
	}
	if !resp.Acknowledged {
		return fmt.Errorf("[%s] index update mapping failed", name)
	}
	return nil
}

// GetMapping 通过索引名称获取 mapping
func (e *Elastic) GetMapping(ctx context.Context, name string) (map[string]interface{}, error) {
	oldMapping, err := e.es.GetMapping().Index(name).Do(ctx)
	if err != nil {
		return nil, err
	}
	if mapping, ok := oldMapping[name].(map[string]interface{})["mappings"]; ok {
		return mapping.(map[string]interface{}), nil
	}
	return nil, fmt.Errorf("[%s] mapping not contains mappings", name)
}

// ReIndex 数据迁移, source : 原始 index  target : 目标 index
func (e *Elastic) ReIndex(ctx context.Context, source, target string) error {
	_, err := e.es.Reindex().SourceIndex(source).DestinationIndex(target).Do(ctx)
	if err != nil {
		return err
	}
	return nil
}
