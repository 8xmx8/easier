package es

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/8xmx8/easier/pkg/logger"
	"github.com/olivere/elastic/v7"
)

// GetDocByID 通过ID获取doc
func (e *Elastic) GetDocByID(ctx context.Context, index, id string, dist interface{}) error {
	doc, err := e.es.Get().Index(index).Id(id).Do(ctx) // nolint
	if err != nil {
		switch {
		case elastic.IsNotFound(err):
			e.loggeer.Info("Document not found", logger.ErrorField(err), logger.MakeField("ID", id))
		case elastic.IsTimeout(err):
			e.loggeer.Error(logger.ErrorES, "Timeout retrieving document", logger.ErrorField(err), logger.MakeField("ID", id))
		case elastic.IsConnErr(err):
			e.loggeer.Error(logger.ErrorES, "Connection problem", logger.ErrorField(err), logger.MakeField("ID", id))
		}
		return err
	}
	return json.Unmarshal(doc.Source, dist)
}

type SearchOption func(*elastic.SearchService)

func SearchWithSort(field string, ascending bool) SearchOption {
	return func(s *elastic.SearchService) {
		s.Sort(field, ascending)
	}
}

// SearchWithPage
// page: 从第0页开始
func SearchWithPage(page, size int) SearchOption {
	if page < 0 {
		page = 0
	}
	from := (page - 1) * size
	return func(s *elastic.SearchService) {
		s.From(from).Size(size)
	}
}

// SearchWithLimit 查询指定条数
// MARK: 不要和`SearchWithPage`同时使用
func SearchWithLimit(limit int) SearchOption {
	return func(s *elastic.SearchService) {
		// limit小于等于0时, 返回默认条数
		if limit <= 0 {
			return
		}
		s.Size(limit)
	}
}

// SearchDocsBySourceQuery 通过构建的原始查询语句进行查询
func (e *Elastic) SearchDocsBySourceQuery(ctx context.Context, index string, query elastic.Query,
	ops ...SearchOption) (*elastic.SearchResult, error) {
	searchService := e.es.Search().Index(index).Query(query)
	for _, op := range ops {
		op(searchService)
	}
	searchResult, err := searchService.TrackTotalHits(true).Do(ctx)
	if err != nil {
		switch {
		case elastic.IsNotFound(err):
			e.loggeer.Infof("Document not found: %v", err)
			return nil, nil
		case elastic.IsTimeout(err):
			e.loggeer.Error(logger.ErrorES, "Timeout retrieving document", logger.ErrorField(err))
		case elastic.IsConnErr(err):
			e.loggeer.Error(logger.ErrorES, "Connection problem", logger.ErrorField(err))
		}
		return nil, err
	}
	return searchResult, nil
}

// SearchDatasByQuery
func (e *Elastic) SearchDocsByTermQuery(ctx context.Context, index string,
	tqs map[string]interface{}, ops ...SearchOption) ([][]byte, error) {
	searchService := e.es.Search().Index(index)
	for k, v := range tqs {
		searchService.Query(elastic.NewTermQuery(k, v))
	}
	for _, op := range ops {
		op(searchService)
	}
	searchResult, err := searchService.TrackTotalHits(true).Do(ctx)
	if err != nil {
		switch {
		case elastic.IsNotFound(err):
			e.loggeer.Infof("Document not found: %v", err)
			return nil, nil
		case elastic.IsTimeout(err):
			e.loggeer.Error(logger.ErrorES, "Timeout retrieving document", logger.ErrorField(err))
		case elastic.IsConnErr(err):
			e.loggeer.Error(logger.ErrorES, "Connection problem", logger.ErrorField(err))
		}
		return nil, err
	}
	if searchResult.Hits.TotalHits.Value < 1 {
		e.loggeer.Info("not found recode", logger.MakeField("index", index))
		return [][]byte{}, nil
	}
	dst := make([][]byte, 0, searchResult.Hits.TotalHits.Value)
	for _, hit := range searchResult.Hits.Hits {
		data, err := hit.Source.MarshalJSON()
		if err != nil {
			e.loggeer.Error(logger.ErrorES, "data marshal", logger.ErrorField(err),
				logger.MakeField("index", hit.Index), logger.MakeField("id", hit.Id))
			continue
		}
		dst = append(dst, data)
	}
	return dst, nil
}

// SearchDocByTermQuery 匹配查询一个doc
// tar: 结果对象指针
// return: docID, err
func (e *Elastic) SearchDocByTermQuery(ctx context.Context, index string,
	tqs map[string]interface{}, tar interface{}, ops ...SearchOption) (string, error) {
	searchService := e.es.Search().Index(index)
	querys := []elastic.Query{}
	for k, v := range tqs {
		querys = append(querys, elastic.NewTermQuery(k, v))
	}
	boolQuery := elastic.NewBoolQuery().Must(querys...)
	searchResult, err := searchService.Query(boolQuery).TrackTotalHits(true).Do(ctx)
	if err != nil {
		switch {
		case elastic.IsNotFound(err):
			e.loggeer.Infof("Document not found: %v", err)
			return "", nil
		case elastic.IsTimeout(err):
			e.loggeer.Error(logger.ErrorES, "Timeout retrieving document", logger.ErrorField(err))
		case elastic.IsConnErr(err):
			e.loggeer.Error(logger.ErrorES, "Connection problem", logger.ErrorField(err))
		}
		return "", err
	}
	if searchResult.Hits.TotalHits.Value < 1 {
		e.loggeer.Info("not found recode", logger.MakeField("index", index))
		return "", nil
	}
	hit := searchResult.Hits.Hits[0]
	if err := json.Unmarshal(hit.Source, tar); err != nil {
		e.loggeer.Error(logger.ErrorES, "doc Unmarshal obj", logger.ErrorField(err), logger.MakeField("index", hit.Index),
			logger.MakeField("id", hit.Id))
		return hit.Id, err
	}
	return hit.Id, nil
}

// SearchDocByQuery 匹配查询一个doc
// tar: 结果对象指针
// return: docID, err
func (e *Elastic) SearchDocByQuery(ctx context.Context, index string,
	query elastic.Query, tar interface{}, ops ...SearchOption) (string, error) {
	searchService := e.es.Search().Index(index)
	searchResult, err := searchService.Query(query).TrackTotalHits(true).Do(ctx)
	if err != nil {
		switch {
		case elastic.IsNotFound(err):
			e.loggeer.Info("Document not found", logger.MakeField(index, "index"))
			return "", nil
		case elastic.IsTimeout(err):
			e.loggeer.Error(logger.ErrorES, "Timeout retrieving document", logger.MakeField(index, "index"), logger.ErrorField(err))
		case elastic.IsConnErr(err):
			e.loggeer.Error(logger.ErrorES, "Connection problem", logger.MakeField(index, "index"), logger.ErrorField(err))
		}
		return "", err
	}
	if searchResult.Hits.TotalHits.Value < 1 {
		// e.loggeer.Info("not found recode", logger.MakeField("index", index))
		return "", nil
	}
	hit := searchResult.Hits.Hits[0]
	if err := json.Unmarshal(hit.Source, tar); err != nil {
		e.loggeer.Error(logger.ErrorES, "doc Unmarshal obj", logger.ErrorField(err), logger.MakeField("index", hit.Index),
			logger.MakeField("id", hit.Id))
		return hit.Id, err
	}
	return hit.Id, nil
}

// nolint
type Aggregation struct {
	Name string
	Agg  elastic.Aggregation
}

// AggregationBySourceQuery 通过构建的原始查询语句进行聚合统计查询
func (e *Elastic) AggregationBySourceQuery(ctx context.Context, index string, query elastic.Query, agg *Aggregation,
	ops ...SearchOption) (*elastic.SearchResult, error) {
	searchService := e.es.Search().Index(index).Query(query).TrackTotalHits(true).Aggregation(agg.Name, agg.Agg).Size(0)
	for _, op := range ops {
		op(searchService)
	}
	searchResult, err := searchService.Do(ctx)
	if err != nil {
		switch {
		case elastic.IsNotFound(err):
			e.loggeer.Infof("Document not found: %v", err)
			return nil, nil
		case elastic.IsTimeout(err):
			e.loggeer.Error(logger.ErrorES, "Timeout retrieving document", logger.ErrorField(err))
		case elastic.IsConnErr(err):
			e.loggeer.Error(logger.ErrorES, "Connection problem", logger.ErrorField(err))
		}
		return nil, err
	}
	return searchResult, nil
}

func (e *Elastic) QueryCountByTime(ctx context.Context, index string, query elastic.Query) (int64, error) {
	searchService := e.es.Search().Index(index).Query(query).Size(0)
	searchResult, err := searchService.TrackTotalHits(true).Do(ctx)
	if err != nil {
		return 0, err
	}
	return searchResult.Hits.TotalHits.Value, nil
}

// AggregateFieldCount
func (e *Elastic) QueryMapsByTime(ctx context.Context, index, filed string, query elastic.Query) (map[string]int64, error) {
	agg := elastic.NewTermsAggregation().Field(filed).Size(1000) // 根据需要调整 Size
	searchService := e.es.Search().Index(index).Query(query).Size(0).Aggregation(fmt.Sprintf("%s_count", filed), agg)

	searchResult, err := searchService.TrackTotalHits(true).Do(ctx)
	if err != nil {
		return nil, err
	}

	aggResult, found := searchResult.Aggregations.Terms(fmt.Sprintf("%s_count", filed))
	if !found {
		return nil, fmt.Errorf("%s_counts aggregation not found", filed)
	}

	resultCounts := make(map[string]int64)
	for _, bucket := range aggResult.Buckets {
		resultCounts[bucket.Key.(string)] = bucket.DocCount
	}
	return resultCounts, nil
}

// DoubleAgg
func (e *Elastic) QueryDoubleAggByTime(ctx context.Context, index, field, ntFiled string,
	query elastic.Query) (map[string]map[string]int64, error) {
	agg := elastic.NewTermsAggregation().Field(field).SubAggregation("status_count",
		elastic.NewTermsAggregation().Field(ntFiled))
	searchResult, err := e.es.Search().
		Index(index).
		Query(query).
		Aggregation("features_count", agg).
		Size(0).
		Do(ctx)

	if err != nil {
		return nil, err
	}
	finalResult := make(map[string]map[string]int64)
	if featuresAgg, found := searchResult.Aggregations.Terms("features_count"); found {
		for _, featureBucket := range featuresAgg.Buckets {
			feature := featureBucket.Key.(string)
			statusAgg, found := featureBucket.Aggregations.Terms("status_count")
			if !found {
				continue
			}
			statusMap := make(map[string]int64)
			for _, statusBucket := range statusAgg.Buckets {
				status := statusBucket.Key.(string)
				count := statusBucket.DocCount
				statusMap[status] = count
			}
			finalResult[feature] = statusMap
		}
	}
	return finalResult, nil
}

func (e *Elastic) QueryDocByOneQuery(ctx context.Context, index string, query elastic.Query) ([]byte, error) {
	searchResult, err := e.es.Search().Index(index).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	if len(searchResult.Hits.Hits) == 0 {
		return nil, err
	}
	return searchResult.Hits.Hits[0].Source, nil
}

func (e *Elastic) QueryTopXByScroll(ctx context.Context, index string, query elastic.Query, x int) (chanr chan []byte, err error) {
	scrollService := e.es.Scroll(index).Query(query).Sort("c_firstAccessTime", true).Size(1000)
	chanr = make(chan []byte, 100)
	totalHits := 0
	go func() {
		defer close(chanr)
		for {
			results, err := scrollService.Do(ctx)
			if totalHits >= x {
				break
			}
			if err != nil {
				return
			}

			for _, hit := range results.Hits.Hits {
				var pd []byte
				if pd, err = json.Marshal(hit.Source); err != nil {
					log.Printf("Failed to unmarshal document: %v", err)
					continue
				}
				select {
				case chanr <- pd:
				case <-ctx.Done():
					return // Ensure graceful shutdown
				}
				totalHits++

				if totalHits >= x {
					break
				}
			}
		}
	}()
	return chanr, nil
}
