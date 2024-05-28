package influxdb

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeletePoint(t *testing.T) {
	ctx := context.Background()
	config := Config{
		Host:     "10.10.11.133:8086",
		Username: "admin",
		Passwd:   "admin123",
		Token:    "B_KX-8VNR8SjcCT0YimUQmVBTudroNHDzjrKiwkP5QmKoS8rtbmkNHLUQsTkYrD4zTNAPW5MS1xr7bnYv0NwvA== ",
		Org:      "eyfm",
		Bucket:   "test",
	}
	client, err := NewInfluxDB(&config)
	assert.NoError(t, err)
	err = client.DeletePoint(ctx, config.Org, config.Bucket)
	assert.NoError(t, err)
}
