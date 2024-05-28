package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	addr = "172.16.20.30:6379"
	db   = 15
)

type MainClientSuite struct {
	suite.Suite
	redis Redis
	ctx   context.Context
}

func Test_MainSuite(t *testing.T) {
	s := &MainClientSuite{
		ctx: context.Background(),
	}
	redis, err := NewClient(s.ctx, addr, db)
	assert.NoError(t, err)
	s.redis = redis
	suite.Run(t, s)
}

func (s *MainClientSuite) BeforeTest(suiteName, testName string) {
	assert.NoError(s.T(), s.redis.(*Client).client.FlushDB().Err())
}

func (s *MainClientSuite) Test_BaseOpreation() {
	convey.Convey("Test_BaseOpreation", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainClientSuite", "BaseOpreation")
		})
		convey.Convey("IsExist", func() {
			fflag := s.redis.IsExist(s.ctx, "hello")
			convey.So(fflag, convey.ShouldBeFalse)
			err := s.redis.SetStr(s.ctx, "hello", "hello world")
			convey.So(err, convey.ShouldBeEmpty)
			tflag := s.redis.IsExist(s.ctx, "hello")
			convey.So(tflag, convey.ShouldBeTrue)
		})
		convey.Convey("Del", func() {
			convey.So(s.redis.Del(s.ctx, "world"), convey.ShouldBeEmpty)
			convey.So(s.redis.SetStr(s.ctx, "world", "All in well"), convey.ShouldBeEmpty)
			convey.So(s.redis.IsExist(s.ctx, "world"), convey.ShouldBeTrue)
			convey.So(s.redis.Del(s.ctx, "world"), convey.ShouldBeEmpty)
			convey.So(s.redis.IsExist(s.ctx, "world"), convey.ShouldBeFalse)
		})
	})
}

func (s *MainClientSuite) Test_String() {
	convey.Convey("Test_String", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainClientSuite", "Test_String")
		})
		convey.Convey("normal_String", func() {
			err := s.redis.SetStr(s.ctx, "hello", "hello world")
			convey.So(err, convey.ShouldBeEmpty)
			var v string
			err = s.redis.GetMixed(s.ctx, "hello", &v)
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(v, convey.ShouldEqual, "hello world")
			convey.So(fmt.Sprintf("%T", v), convey.ShouldEqual, "string")
		})
		convey.Convey("SetExpire_String", func() {
			err := s.redis.SetStr(s.ctx, "helloExpire", "hello world for Expire")
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(s.redis.SetExpire(s.ctx, "helloExpire", time.Second), convey.ShouldBeEmpty)
			<-time.NewTimer(500 * time.Millisecond).C
			convey.So(s.redis.IsExist(s.ctx, "helloExpire"), convey.ShouldBeTrue)
			<-time.NewTimer(time.Second).C
			convey.So(s.redis.IsExist(s.ctx, "helloExpire"), convey.ShouldBeFalse)
		})
		convey.Convey("withTTL_String", func() {
			err := s.redis.SetStrTTL(s.ctx, "helloTTL", "hello world for TTL", time.Second)
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(s.redis.SetExpire(s.ctx, "helloTTL", time.Second), convey.ShouldBeEmpty)
			<-time.NewTimer(500 * time.Millisecond).C
			convey.So(s.redis.IsExist(s.ctx, "helloTTL"), convey.ShouldBeTrue)
			<-time.NewTimer(time.Second).C
			convey.So(s.redis.IsExist(s.ctx, "helloTTL"), convey.ShouldBeFalse)
		})
		convey.Convey("int_String", func() {
			err := s.redis.SetStr(s.ctx, "hello10086", "10086")
			convey.So(err, convey.ShouldBeEmpty)
			var v string
			err = s.redis.GetMixed(s.ctx, "hello10086", &v)
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(v, convey.ShouldEqual, "10086")
			convey.So(fmt.Sprintf("%T", v), convey.ShouldEqual, "string")
		})
		convey.Convey("listJson_String", func() {
			v := []interface{}{"hello", 123123, []int{1, 2, 3, 4, 5}}
			vb, err := json.Marshal(v)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.redis.SetStr(s.ctx, "hello_json", string(vb))
			convey.So(err, convey.ShouldBeEmpty)
			var resv string
			err = s.redis.GetMixed(s.ctx, "hello_json", &resv)
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(resv, convey.ShouldEqual, string(vb))
			convey.So(fmt.Sprintf("%T", resv), convey.ShouldEqual, "string")
			var tar []interface{}
			err = json.Unmarshal([]byte(resv), &tar)
			convey.So(err, convey.ShouldBeEmpty)
			// convey.So(tar, convey.ShouldResemble, v)
		})
		convey.Convey("structJson_String", func() {
			type Person struct {
				Name  string
				Age   int
				Addr  string
				Hobby []string
			}
			p := &Person{
				Name:  "tom",
				Age:   18,
				Addr:  "北京·东城",
				Hobby: []string{"计算机", "音乐"},
			}
			vb, err := json.Marshal(p)
			convey.So(err, convey.ShouldBeEmpty)
			err = s.redis.SetStr(s.ctx, "hello_person", string(vb))
			convey.So(err, convey.ShouldBeEmpty)
			var resv string
			err = s.redis.GetMixed(s.ctx, "hello_person", &resv)
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(resv, convey.ShouldEqual, string(vb))
			convey.So(fmt.Sprintf("%T", resv), convey.ShouldEqual, "string")
			var tar Person
			err = json.Unmarshal([]byte(resv), &tar)
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(tar, convey.ShouldResemble, *p)
		})
	})
}

func (s *MainClientSuite) Test_Hash() {
	convey.Convey("Test_Hash", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainClientSuite", "Test_Hash")
		})
		convey.Convey("normal_Hash", func() {
			hmap := map[string]interface{}{
				"Name": "tom",
				"Age":  18,
				"Addr": "北京·东城",
				// "Hobby": []string{"计算机", "音乐"},
			}
			convey.So(s.redis.SetHash(s.ctx, "hello_hash", hmap), convey.ShouldBeEmpty)

			tarV, err := s.redis.GetHashField(s.ctx, "hello_hash", "Age")
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(tarV, convey.ShouldEqual, "18")

			// TODO: @zcf redis 的Hash不支持 层级结构
			// tarH, err := s.redis.GetHashField(s.ctx, "hello_hash", "Hobby")
			// convey.So(err, convey.ShouldBeEmpty)
			// hb, err := json.Marshal(hmap["Hobby"])
			// convey.So(err, convey.ShouldBeEmpty)
			// convey.So(tarH, convey.ShouldEqual, string(hb))

		})
		convey.Convey("SetExpire_Hash", func() {
			p := map[string]interface{}{
				"Name": "张三",
				"Age":  20,
				"Addr": "北京·西城",
			}
			convey.So(s.redis.SetHash(s.ctx, "hello_p1", p), convey.ShouldBeEmpty)
			convey.So(s.redis.SetExpire(s.ctx, "hello_p1", time.Second), convey.ShouldBeEmpty)
			<-time.NewTimer(500 * time.Millisecond).C
			convey.So(s.redis.IsExist(s.ctx, "hello_p1"), convey.ShouldBeTrue)
			<-time.NewTimer(time.Second).C
			convey.So(s.redis.IsExist(s.ctx, "hello_p1"), convey.ShouldBeFalse)
		})
	})
}

func (s *MainClientSuite) Test_List() {
	convey.Convey("Test_List", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainClientSuite", "Test_List")
		})
		convey.Convey("normal_List", func() {
			ll := []interface{}{123, "hello", 3.14159, 66788}
			convey.So(s.redis.PushList(s.ctx, "hello_list", ll...), convey.ShouldBeEmpty)

			var getL []string
			err := s.redis.GetMixed(s.ctx, "hello_list", &getL)
			convey.So(err, convey.ShouldBeEmpty)
			// convey.So(getL, convey.ShouldEqual, ll)
			for i, v := range getL {
				convey.So(v, convey.ShouldEqual, fmt.Sprint(ll[i]))
			}
		})
		convey.Convey("push_List", func() {
			ll := []interface{}{123, "hello", 3.14159, 66788}
			convey.So(s.redis.PushList(s.ctx, "push_list", ll...), convey.ShouldBeEmpty)
			len1, err := s.redis.LenList(s.ctx, "push_list")
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(len1, convey.ShouldEqual, len(ll))

			convey.So(s.redis.PushList(s.ctx, "push_list", 0.618, 1<<10), convey.ShouldBeEmpty)
			convey.So(err, convey.ShouldBeEmpty)
			len2, err := s.redis.LenList(s.ctx, "push_list")
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(len2, convey.ShouldEqual, len(ll)+2)

		})
		convey.Convey("pop_List", func() {
			ll := []interface{}{123, "hello", 3.14159, 66788}
			convey.So(s.redis.PushList(s.ctx, "pop_list", ll...), convey.ShouldBeEmpty)
			var item1 int
			convey.So(s.redis.PopList(s.ctx, "pop_list", &item1), convey.ShouldBeEmpty)
			convey.So(item1, convey.ShouldEqual, 123)
			len1, err := s.redis.LenList(s.ctx, "pop_list")
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(len1, convey.ShouldEqual, len(ll)-1)

			var item2 string
			convey.So(s.redis.PopList(s.ctx, "pop_list", &item2), convey.ShouldBeEmpty)
			convey.So(item2, convey.ShouldEqual, "hello")
			len2, err := s.redis.LenList(s.ctx, "pop_list")
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(len2, convey.ShouldEqual, len(ll)-2)

			var item3 float32
			convey.So(s.redis.PopList(s.ctx, "pop_list", &item3), convey.ShouldBeEmpty)
			convey.So(item3, convey.ShouldEqual, 3.14159)
			len3, err := s.redis.LenList(s.ctx, "pop_list")
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(len3, convey.ShouldEqual, len(ll)-3)
		})
	})
}

func (s *MainClientSuite) Test_Set() {
	convey.Convey("Test_Set", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainClientSuite", "Test_Set")
		})
		convey.Convey("normal_set", func() {
			ll := []interface{}{123, "hello", 3.14159, 66788, "hello", 66788}
			convey.So(s.redis.AddSet(s.ctx, "helloSet", ll...), convey.ShouldBeEmpty)
			var Vset []string
			convey.So(s.redis.GetMixed(s.ctx, "helloSet", &Vset), convey.ShouldBeEmpty)
			convey.So(len(Vset), convey.ShouldEqual, 4)
		})
		convey.Convey("check_set_member", func() {
			ll := []interface{}{123, "hello", 3.14159, 66788, "hello", 66788}
			convey.So(s.redis.AddSet(s.ctx, "checkSet", ll...), convey.ShouldBeEmpty)
			res, err := s.redis.CheckSetMember(s.ctx, "checkSet", "hello")
			convey.So(err, convey.ShouldBeNil)
			convey.So(res, convey.ShouldBeTrue)
		})
		convey.Convey("RemEle_set", func() {
			ll := []interface{}{123, "hello", 3.14159, 66788, "hello", 66788}
			convey.So(s.redis.AddSet(s.ctx, "RemEleSet", ll...), convey.ShouldBeEmpty)
			convey.So(s.redis.RemSetEle(s.ctx, "RemEleSet", "hello", 66788), convey.ShouldBeEmpty)
			var Vset []string
			convey.So(s.redis.GetMixed(s.ctx, "RemEleSet", &Vset), convey.ShouldBeEmpty)
			convey.So(len(Vset), convey.ShouldEqual, 2)
		})
	})
}

func (s *MainClientSuite) Test_ZSet() {
	convey.Convey("Test_ZSet", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainClientSuite", "Test_ZSet")
		})
		convey.Convey("normal_zset", func() {
			zmap := []*ZSetMember{
				{Score: 3.14, Member: "hello"},
				{Score: 0.618, Member: "world"},
				{Score: 1, Member: "tan"},
				{Score: -0.5, Member: "sin"},
			}
			convey.So(s.redis.AddZSet(s.ctx, "hello_ZSet", zmap...), convey.ShouldBeEmpty)
			var vZset []string
			convey.So(s.redis.GetMixed(s.ctx, "hello_ZSet", &vZset), convey.ShouldBeEmpty)
			convey.So(vZset, convey.ShouldResemble, []string{"hello", "tan", "world", "sin"})

			zCard, err := s.redis.CardZSet(s.ctx, "hello_ZSet")
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(zCard, convey.ShouldEqual, len(vZset))
		})
		convey.Convey("RemMembers_zset", func() {
			zmap := []*ZSetMember{
				{Score: 3.14, Member: "hello"},
				{Score: 0.618, Member: "world"},
				{Score: 1, Member: "tan"},
				{Score: -0.5, Member: "sin"},
			}
			convey.So(s.redis.AddZSet(s.ctx, "rem_ZSet", zmap...), convey.ShouldBeEmpty)
			convey.So(s.redis.RemMembersZSet(s.ctx, "rem_ZSet", "tan", "sin"), convey.ShouldBeEmpty)
			var vZset []string
			convey.So(s.redis.GetMixed(s.ctx, "rem_ZSet", &vZset), convey.ShouldBeEmpty)
			convey.So(vZset, convey.ShouldResemble, []string{"hello", "world"})
		})
		convey.Convey("WithScore_zset", func() {
			zmap := []*ZSetMember{
				{Score: 3.14, Member: "hello"},
				{Score: 0.618, Member: "world"},
				{Score: 1, Member: "tan"},
				{Score: -0.5, Member: "sin"},
				{Score: 100, Member: 123},
			}
			convey.So(s.redis.AddZSet(s.ctx, "ws_ZSet", zmap...), convey.ShouldBeEmpty)
			res, err := s.redis.MembersWithScoreZSet(s.ctx, "ws_ZSet")
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(res, convey.ShouldResemble, []*ZSetMember{
				{Score: 100, Member: "123"},
				{Score: 3.14, Member: "hello"},
				{Score: 1, Member: "tan"},
				{Score: 0.618, Member: "world"},
				{Score: -0.5, Member: "sin"},
			})
		})

	})
}
