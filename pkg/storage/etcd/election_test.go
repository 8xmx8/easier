package etcd

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	serviceName = "/test_service"
)

func TestETCDElection(t *testing.T) {
	ctx := context.Background()
	// new ETCD client
	cc, err := NewETCD(ctx, []string{host})
	if !assert.NoError(t, err) {
		return
	}
	// TODO: 未完, 后期再补
	fmt.Println(cc)
}
