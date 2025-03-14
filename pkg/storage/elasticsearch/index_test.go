package es

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testIndex = "test_data_sl"
	addrs     = []string{"http://0.0.0.0:9200"}
	user      = ""
	passwd    = ""
)

func TestCreateIndex(t *testing.T) {
	ctx := context.Background()
	es, err := NewElastic(ctx, &Config{Addrs: addrs, User: user, Pwd: passwd})
	assert.NoError(t, err)
	mapping := map[string]interface{}{
		"properties": map[string]interface{}{
			"ip": map[string]interface{}{
				"type":  "ip",
				"store": true,
			},
			"port": map[string]interface{}{
				"type": "long",
			},
			"lastupdatetime": map[string]interface{}{
				"format": "YYYY-MM-dd HH:mm:ss",
				"type":   "date",
			},
		},
	}
	assert.NoError(t, es.CreateIndex(ctx, testIndex, mapping))
}

func TestDeleteIndex(t *testing.T) {
	ctx := context.Background()
	es, err := NewElastic(ctx, &Config{Addrs: addrs})
	assert.NoError(t, err)
	assert.NoError(t, es.DeleteIndex(ctx, testIndex))
}

func TestUpdateMappingByReIndex(t *testing.T) {

	ctx := context.Background()
	es, err := NewElastic(ctx, &Config{Addrs: addrs, User: user, Pwd: passwd})
	assert.NoError(t, err)

	mapping1 := map[string]interface{}{
		"properties": map[string]interface{}{
			"ping": map[string]interface{}{
				"dynamic": "true",
				"properties": map[string]interface{}{
					"address": map[string]interface{}{
						"type": "keyword",
					},
					"ip": map[string]interface{}{
						"type": "keyword",
					},
					"avg": map[string]interface{}{
						"type": "long",
					},
					"connectivity": map[string]interface{}{
						"type": "boolean",
					},
					"duplicates": map[string]interface{}{
						"type": "long",
					},
					"loss": map[string]interface{}{
						"type": "double",
					},
					"max": map[string]interface{}{
						"type": "long",
					},
					"min": map[string]interface{}{
						"type": "long",
					},
					"received": map[string]interface{}{
						"type": "long",
					},
					"stddev": map[string]interface{}{
						"type": "long",
					},
					"transmitted": map[string]interface{}{
						"type": "long",
					},
					"provincesName": map[string]interface{}{
						"type": "keyword",
					},
				},
			},
			"dns": map[string]interface{}{
				"dynamic": "true",
				"properties": map[string]interface{}{
					"rtt": map[string]interface{}{
						"type": "long",
					},
					"data": map[string]interface{}{
						"dynamic": "true",
						"properties": map[string]interface{}{
							"class": map[string]interface{}{
								"type": "keyword",
							},
							"name": map[string]interface{}{
								"type": "keyword",
							},
							"rrtype": map[string]interface{}{
								"type": "keyword",
							},
							"target": map[string]interface{}{
								"type": "keyword",
							},
							"ttl": map[string]interface{}{
								"type": "long",
							},
						},
					},
				},
			},
			"web": map[string]interface{}{
				"dynamic": "true",
				"properties": map[string]interface{}{
					"connectDuration": map[string]interface{}{
						"type": "long",
					},
					"firstByteDuration": map[string]interface{}{
						"type": "long",
					},
					"status": map[string]interface{}{
						"type": "long",
					},
					"contentDuration": map[string]interface{}{
						"type": "long",
					},
				},
			},
			"errorMessage": map[string]interface{}{
				"dynamic":    "true",
				"properties": map[string]interface{}{},
			},
			"isResultOK": map[string]interface{}{
				"type": "boolean",
			},
			"beginAt": map[string]interface{}{
				"format": "yyyy-MM-dd HH:mm:ss",
				"type":   "date",
			},
			"endAt": map[string]interface{}{
				"format": "yyyy-MM-dd HH:mm:ss",
				"type":   "date",
			},
			"storageAt": map[string]interface{}{
				"format": "yyyy-MM-dd HH:mm:ss",
				"type":   "date",
			},
			"nodeID": map[string]interface{}{
				"type":  "keyword",
				"index": true,
			},
			"taskID": map[string]interface{}{
				"type":  "keyword",
				"index": true,
			},
			"targetHost": map[string]interface{}{
				"type":  "keyword",
				"index": true,
			},
			"parentTaskID": map[string]interface{}{
				"type":  "keyword",
				"index": true,
			},
		},
	}

	assert.NoError(t, es.UpdateMappingByReIndex(ctx, "certdata-netquality.1681272235", mapping1))
}

func TestUpdateMapping(t *testing.T) {
	ctx := context.Background()
	es, err := NewElastic(ctx, &Config{Addrs: addrs, User: user, Pwd: passwd})
	assert.NoError(t, err)

	mapping := map[string]interface{}{
		"properties": map[string]interface{}{
			"ping": map[string]interface{}{
				"dynamic": "true",
				"properties": map[string]interface{}{
					"address": map[string]interface{}{
						"type": "keyword",
					},
					"ip": map[string]interface{}{
						"type": "keyword",
					},
					"avg": map[string]interface{}{
						"type": "long",
					},
					"connectivity": map[string]interface{}{
						"type": "boolean",
					},
					"duplicates": map[string]interface{}{
						"type": "long",
					},
					"loss": map[string]interface{}{
						"type": "double",
					},
					"max": map[string]interface{}{
						"type": "long",
					},
					"min": map[string]interface{}{
						"type": "long",
					},
					"received": map[string]interface{}{
						"type": "long",
					},
					"stddev": map[string]interface{}{
						"type": "long",
					},
					"transmitted": map[string]interface{}{
						"type": "long",
					},
					"provincesName": map[string]interface{}{
						"type": "keyword",
					},
				},
			},
			"dns": map[string]interface{}{
				"dynamic": "true",
				"properties": map[string]interface{}{
					"rtt": map[string]interface{}{
						"type": "long",
					},
					"data": map[string]interface{}{
						"dynamic": "true",
						"properties": map[string]interface{}{
							"class": map[string]interface{}{
								"type": "keyword",
							},
							"name": map[string]interface{}{
								"type": "keyword",
							},
							"rrtype": map[string]interface{}{
								"type": "keyword",
							},
							"target": map[string]interface{}{
								"type": "keyword",
							},
							"ttl": map[string]interface{}{
								"type": "long",
							},
						},
					},
				},
			},
			"web": map[string]interface{}{
				"dynamic": "true",
				"properties": map[string]interface{}{
					"connectDuration": map[string]interface{}{
						"type": "long",
					},
					"firstByteDuration": map[string]interface{}{
						"type": "long",
					},
					"status": map[string]interface{}{
						"type": "long",
					},
					"contentDuration": map[string]interface{}{
						"type": "long",
					},
				},
			},
			"errorMessage": map[string]interface{}{
				"dynamic":    "true",
				"properties": map[string]interface{}{},
			},
			"isResultOK": map[string]interface{}{
				"type": "boolean",
			},
			"beginAt": map[string]interface{}{
				"format": "yyyy-MM-dd HH:mm:ss",
				"type":   "date",
			},
			"endAt": map[string]interface{}{
				"format": "yyyy-MM-dd HH:mm:ss",
				"type":   "date",
			},
			"storageAt": map[string]interface{}{
				"format": "yyyy-MM-dd HH:mm:ss",
				"type":   "date",
			},
			"nodeID": map[string]interface{}{
				"type":  "keyword",
				"index": true,
			},
			"taskID": map[string]interface{}{
				"type":  "keyword",
				"index": true,
			},
			"targetHost": map[string]interface{}{
				"type":  "keyword",
				"index": true,
			},
			"parentTaskID": map[string]interface{}{
				"type":  "keyword",
				"index": true,
			},
		},
	}

	assert.NoError(t, es.UpdateMapping(ctx, "certdata-netquality.1681272235", mapping))

}
