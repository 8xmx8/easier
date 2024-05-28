package etcd

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MainLockSuite struct {
	suite.Suite
	etcd *Client
	ctx  context.Context
}

func Test_MainLockSuite(t *testing.T) {
	ctx := context.Background()
	etcd, err := NewETCD(ctx, []string{"172.16.20.30:2379"})
	assert.Nil(t, err)
	s := &MainLockSuite{
		ctx:  ctx,
		etcd: etcd,
	}
	suite.Run(t, s)
}

func (s *MainLockSuite) BeforeTest(suiteName, testName string) {
	for key, _ := range s.etcd.locks {
		s.etcd.DestroyLock(s.ctx, key)
	}
}
func (s *MainLockSuite) Test_Lock() {
	convey.Convey("Test_Lock&Unlock", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainLockSuite", "Lock&Unlock")
		})
		convey.Convey("Lock&Unlock", func() {
			err := s.etcd.Lock(s.ctx, "hello")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
