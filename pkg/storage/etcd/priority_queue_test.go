package etcd

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/suite"
)

type MainSuite struct {
	suite.Suite
	queue     *PriorityQueue
	endpoints []string
}

func Test_MainSuite(t *testing.T) {
	s := &MainSuite{
		endpoints: []string{"172.16.20.30:2379"},
	}
	suite.Run(t, s)
}

func (s *MainSuite) BeforeTest(suiteName, testName string) {
	s.queue = nil
}
func (s *MainSuite) Test_Queue() {
	convey.Convey("Test_QueueOneTag", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainSuite", "Test_QueueOneTag")
		})
		queue, err := NewPriorityQueue(context.Background(), s.endpoints, "Test_QueueOneTag")
		s.queue = queue
		convey.So(err, convey.ShouldBeEmpty)

		convey.Convey("push&pop_one", func() {
			value := "hello one"
			err := s.queue.Push(context.TODO(), value, 1)
			convey.So(err, convey.ShouldBeEmpty)
			str, err := s.queue.Pop(context.TODO())
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(str, convey.ShouldEqual, value)
		})
	})

	convey.SkipConvey("Test_QueueSameTag", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainSuite", "Test_QueueSameTag")
		})
		queue, err := NewPriorityQueue(context.Background(), s.endpoints, "Test_QueueSameTag")
		s.queue = queue
		convey.So(err, convey.ShouldBeEmpty)

		convey.Convey("push&pop_same", func() {
			// Push
			err := s.queue.Push(context.TODO(), "hello one", 1)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.queue.Push(context.TODO(), "hello two", 1)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.queue.Push(context.TODO(), "hello three", 1)
			convey.So(err, convey.ShouldBeEmpty)
			// POP
			str1, err := s.queue.Pop(context.TODO())
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(str1, convey.ShouldEqual, "hello one")
			str2, err := s.queue.Pop(context.TODO())
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(str2, convey.ShouldEqual, "hello two")
			str3, err := s.queue.Pop(context.TODO())
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(str3, convey.ShouldEqual, "hello three")
		})
	})

	convey.SkipConvey("Test_QueueSomeTag", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainSuite", "Test_QueueSomeTag")
		})
		queue, err := NewPriorityQueue(context.Background(), s.endpoints, "Test_QueueSomeTag")
		s.queue = queue
		convey.So(err, convey.ShouldBeEmpty)

		convey.Convey("push&pop_some_sort", func() {
			// Push
			err := s.queue.Push(context.TODO(), "hello one", 1)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.queue.Push(context.TODO(), "hello two", 2)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.queue.Push(context.TODO(), "hello three", 3)
			convey.So(err, convey.ShouldBeEmpty)

			// POP
			str1, err := s.queue.Pop(context.TODO())
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(str1, convey.ShouldEqual, "hello one")
			str2, err := s.queue.Pop(context.TODO())
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(str2, convey.ShouldEqual, "hello two")
			str3, err := s.queue.Pop(context.TODO())
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(str3, convey.ShouldEqual, "hello three")
		})

		convey.Convey("push&pop_some_sort_2", func() {
			// Push
			err := s.queue.Push(context.TODO(), "hello three", 3)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.queue.Push(context.TODO(), "hello one", 1)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.queue.Push(context.TODO(), "hello three2", 3)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.queue.Push(context.TODO(), "hello two", 2)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.queue.Push(context.TODO(), "hello one2", 1)
			convey.So(err, convey.ShouldBeEmpty)
			// POP
			str1, err := s.queue.Pop(context.TODO())
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(str1, convey.ShouldEqual, "hello one")
			str2, err := s.queue.Pop(context.TODO())
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(str2, convey.ShouldEqual, "hello one2")
			str3, err := s.queue.Pop(context.TODO())
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(str3, convey.ShouldEqual, "hello two")
			str4, err := s.queue.Pop(context.TODO())
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(str4, convey.ShouldEqual, "hello three")
			str5, err := s.queue.Pop(context.TODO())
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(str5, convey.ShouldEqual, "hello three2")
		})
	})
}

func (s *MainSuite) Test_LenQueue() {
	convey.Convey("Test_LenQueue", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainSuite", "LenQueue")
		})
		queue, err := NewPriorityQueue(context.Background(), s.endpoints, "Test_LenQueue")
		s.queue = queue
		convey.So(err, convey.ShouldBeEmpty)
		convey.Convey("len_queue", func() {
			// Push
			err = s.queue.Push(context.TODO(), "hello three", 3)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.queue.Push(context.TODO(), "hello one", 1)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.queue.Push(context.TODO(), "hello three2", 3)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.queue.Push(context.TODO(), "hello two", 2)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.queue.Push(context.TODO(), "hello one2", 1)
			convey.So(err, convey.ShouldBeEmpty)

			// 执行
			l, err := s.queue.Len(context.TODO())
			convey.So(err, convey.ShouldBeNil)
			convey.So(l, convey.ShouldEqual, 5)
		})
	})
}

func (s *MainSuite) Test_DeleteQueue() {
	convey.Convey("Test_DeleteQueue", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainSuite", "DeleteQueue")
		})
		queue, err := NewPriorityQueue(context.Background(), s.endpoints, "Test_DeleteQueue")
		s.queue = queue
		convey.So(err, convey.ShouldBeEmpty)
		convey.Convey("delete_queue", func() {
			// Push
			err = s.queue.Push(context.TODO(), "hello three", 3)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.queue.Push(context.TODO(), "hello one", 1)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.queue.Push(context.TODO(), "hello three2", 3)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.queue.Push(context.TODO(), "hello two", 2)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.queue.Push(context.TODO(), "hello one2", 1)
			convey.So(err, convey.ShouldBeEmpty)

			// 删除
			queue.Delete(context.TODO())

			// 验证
			l, err := s.queue.Len(context.TODO())
			convey.So(err, convey.ShouldBeNil)
			convey.So(l, convey.ShouldEqual, 0)
		})
	})
}

func (s *MainSuite) Test_GetQueueKeys() {
	convey.Convey("Test_GetQueueKeys", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainSuite", "Test_GetQueueKeys")
		})
		queue, err := NewPriorityQueue(context.Background(), s.endpoints, "Test_GetQueueKeys")
		s.queue = queue
		convey.So(err, convey.ShouldBeEmpty)

		convey.Convey("get_queue_keys", func() {
			// Push
			err = s.queue.Push(context.TODO(), "hello three", 3)
			convey.So(err, convey.ShouldBeEmpty)

			keys, err := s.queue.etcd.GetPriorityQueueList(context.Background(), "Test_GetQueueKey")
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(len(keys), convey.ShouldEqual, 1)
			convey.So("Test_GetQueueKeys", convey.ShouldBeIn, keys)
		})
	})
}
