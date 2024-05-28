package influxdb

import (
	"context"
	"fmt"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"

	"time"
)

type Config struct {
	Host     string `json:"host"`     // InfluxDB address, in the format ip:port
	Username string `json:"username"` // Username for InfluxDB
	Passwd   string `json:"passwd"`   // Password for InfluxDB
	DB       string `json:"db"`       // Database name
	Token    string `json:"token"`    // Token for authentication
	Org      string `json:"org"`      // Organization name
	Bucket   string `json:"bucket"`   // Bucket name
}

type Client struct {
	Client    influxdb2.Client
	WriteAPI  api.WriteAPI
	QueryAPI  api.QueryAPI
	DeleteAPI api.DeleteAPI
}

func NewInfluxDB(config *Config) (*Client, error) {
	// 可以自己封装
	url := fmt.Sprintf("http://%s", config.Host)
	client := influxdb2.NewClient(url, config.Token)
	writeAPI := client.WriteAPI(config.Org, config.Bucket)
	queryAPI := client.QueryAPI(config.Org)
	deleteAPI := client.DeleteAPI()

	return &Client{
		Client:    client,
		WriteAPI:  writeAPI,
		QueryAPI:  queryAPI,
		DeleteAPI: deleteAPI,
	}, nil
}

func (c *Client) Close() {
	c.Client.Close()
}

func (c *Client) WritePoint(ctx context.Context, p *write.Point) error {
	c.WriteAPI.WritePoint(p)
	return nil
}
func (c *Client) NewPointIn(ctx context.Context, measurement string,
	tags map[string]string,
	fields map[string]interface{},
	ts time.Time) (*write.Point, error) {
	return write.NewPoint(measurement, tags, fields, ts), nil
}
func (c *Client) DeletePoint(ctx context.Context, org, bucket string) error {
	return c.DeleteAPI.DeleteWithName(ctx, org, bucket, time.UnixMicro(0), time.Now(), "")
}

func (c *Client) QueryCountDaymapByTime(ctx context.Context, query string) (map[string]int64, error) {
	results, err := c.QueryAPI.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	countsByDay := make(map[string]int64)
	for results.Next() {
		if stopTime := results.Record().ValueByKey("_start"); stopTime != nil {
			day := stopTime.(time.Time).Format("2006-01-02")
			if value, ok := results.Record().Value().(int64); ok {
				countsByDay[day] = value
			}
		}
	}
	if results.Err() != nil {
		return nil, results.Err()
	}
	return countsByDay, nil
}

func (c *Client) QueryCountByTime(ctx context.Context, query string) (count int64, err error) {
	results, err := c.QueryAPI.Query(ctx, query)
	if err != nil {
		return 0, err
	}
	for results.Next() {
		if value, ok := results.Record().Value().(int64); ok {
			count += value
		}
	}
	if results.Err() != nil {
		return 0, results.Err()
	}
	return count, nil
}
