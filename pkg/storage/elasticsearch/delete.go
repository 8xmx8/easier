package es

import (
	"context"

	"github.com/olivere/elastic/v7"
)

// DeleteDocByQuery 根据条件删除文档
func (e *Elastic) DeleteDocByQuery(ctx context.Context, index string, query elastic.Query) (int64, error) {
	rsp, err := e.es.DeleteByQuery(index).Query(query).Refresh("true").Do(ctx)
	if err != nil {
		return 0, err
	}
	return rsp.Deleted, nil
}

// DeleteDocByID 根据索引ID删除文档
func (e *Elastic) DeleteDocByID(ctx context.Context, index, id string) error {
	_, err := e.es.Delete().Index(index).Id(id).Refresh("true").Do(ctx)
	return err
}
