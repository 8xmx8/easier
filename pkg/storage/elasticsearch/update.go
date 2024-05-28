package es

import "context"

// UpdateCoverDocByID 覆盖更新doc
func (e *Elastic) UpdateCoverDocByID(ctx context.Context, index, id string, doc interface{}) error {
	_, err := e.es.Update().Index(index).Id(id).Doc(doc).DocAsUpsert(true).Do(ctx)
	return err
}

// DocUpdateByID 根据docID实现局部更新
func (e *Elastic) DocPartialUpdatesByID(ctx context.Context, index, docID string, data map[string]any) error {
	_, err := e.es.Update().Index(index).Id(docID).Doc(data).Do(ctx)
	if err != nil {
		return err
	}
	return nil
}
