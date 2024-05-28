package bloomfilter

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MainSuite struct {
	suite.Suite
	filter *BloomFilter
	ctx    context.Context
}

func Test_NetMainSuite(t *testing.T) {
	ctx := context.Background()
	// 1w个样本, 万分之一的误差
	filter, err := NewBloom(ctx, 1000000, 0.0001)
	assert.NoError(t, err)
	s := &MainSuite{
		ctx:    ctx,
		filter: filter,
	}
	suite.Run(t, s)
}

func (s *MainSuite) BeforeTest(suiteName, testName string) {
	s.filter.ClearAll()
}

func (s *MainSuite) Test_AddAndTest() {
	convey.Convey("Test_AddAndTest", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainSuite", "AddAndTest")
		})
		// 构造测试用例
		exampleList := make([]string, 0, 1000)
		for i := 0; i < 1000; i++ {
			key := faker.ChineseName()
			exampleList = append(exampleList, key)
			s.filter.Add([]byte(key))
		}
		convey.Convey("存在测试", func() {
			for _, key := range exampleList {
				isExit := s.filter.Test([]byte(key))
				convey.So(isExit, convey.ShouldBeTrue)
			}
		})
		convey.Convey("不存在测试", func() {
			for i := 0; i < 1000; i++ {
				key := faker.DomainName()
				isExit := s.filter.Test([]byte(key))
				convey.So(isExit, convey.ShouldBeFalse)
			}
		})
	})
}

func (s *MainSuite) Test_ByFile() {
	convey.Convey("Test_ByFile", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainSuite", "ByFile")
		})
		// 构造测试用例
		exampleList := make([]string, 0, 1000)
		for i := 0; i < 1000; i++ {
			key := faker.ChineseName()
			exampleList = append(exampleList, key)
			s.filter.Add([]byte(key))
		}
		convey.SkipConvey("写入文件", func() {
			err := s.filter.DownloadToFile(s.ctx, "aaa")
			convey.So(err, convey.ShouldBeEmpty)
			// 验证文件存在
		})
		convey.Convey("从文件读取", func() {
			err := s.filter.LoadByFile(s.ctx, "aaa")
			convey.So(err, convey.ShouldBeEmpty)
		})
	})
}
