package es

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateDoc(t *testing.T) {
	ctx := context.Background()
	es, err := NewElastic(ctx, &Config{Addrs: addrs})
	assert.NoError(t, err)
	//data := map[string]interface{}{
	//	"ip":             "127.0.0.1",
	//	"port":           10086,
	//	"lastupdatetime": time.Now().Format("2006-01-02 15:04:05"),
	//}
	assert.NoError(t, es.CreateDoc(ctx, "wa_wa1c_eywzfx_db_t_ads_eywz_bpp", LoadMapping("wa_wa1c_eywzfx_db_t_ads_eywz_bpp.json")))
}

func TestDeleteDocByQuery(t *testing.T) {
	ctx := context.Background()
	es, err := NewElastic(ctx, &Config{Addrs: addrs})
	assert.NoError(t, err)
	query := elastic.NewBoolQuery()
	query = query.MustNot(elastic.NewTermQuery("taskID", "123"))
	query = query.Must(elastic.NewTermQuery("targetHost", "futurearchitect.biz"))
	d, err := es.DeleteDocByQuery(ctx, "certdata-domainsubdomain", query)
	fmt.Println(d)
	assert.NoError(t, err)
}

func TestDocIsExist(t *testing.T) {
	ctx := context.Background()
	es, err := NewElastic(ctx, &Config{Addrs: addrs})
	assert.NoError(t, err)
	isExist := es.DocIsExist(ctx, "certdata-blockchainnodedetect", "82.223.23.73:9108")
	assert.True(t, isExist)
	notExist := es.DocIsExist(ctx, "certdata-blockc", "82.223.23.73:9108")
	assert.False(t, notExist)
}
