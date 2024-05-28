package es

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	searchUser  = ""
	searchPwd   = ""
	searchAddrs = []string{"http://10.11.12.35:9200"}
)

type MainSuite struct {
	suite.Suite
	esClient *Elastic
	ctx      context.Context
}

func Test_MainSuite(t *testing.T) {
	s := &MainSuite{}
	suite.Run(t, s)
}

func (s *MainSuite) BeforeTest(suiteName, testName string) {
	s.ctx = context.Background()
	esclient, err := NewElastic(s.ctx, &Config{
		User:  searchUser,
		Pwd:   searchPwd,
		Addrs: searchAddrs,
	})
	assert.NoError(s.T(), err)
	s.esClient = esclient
}

func (s *MainSuite) Test_GetDocByID() {
	convey.SkipConvey("Test_GetDocByID", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainSuite", "Test_GetDocByID")
		})
		convey.Convey("normal_GetDocByID", func() {
			var res map[string]interface{}
			err := s.esClient.GetDocByID(s.ctx, "analysis-assets_poc", "zv1h-YMBRzVnG-Ha1e8n", &res)
			convey.So(err, convey.ShouldBeNil)
			convey.So(res, convey.ShouldResemble, map[string]interface{}{
				"target":       "directgenerate.name",
				"scanAt":       "2022-10-21 15:11:20",
				"fileName":     "wordpress_wp_license_file_read.json",
				"hostInfo":     "http://124.236.103.51",
				"name":         "wordpress内容管理系统后门文件wp-license.php页面file参数任意文件读取",
				"vulnerable":   true,
				"data":         "",
				"lastResponse": "",
				"level":        "1",
				"method":       "DELETE",
				"path":         "https://www.districtmesh.name/virtual/sticky/technologies/sexy",
				"vulURL":       "http://www.legacyorchestrate.biz/synergistic/maximize/infrastructures/best-of-breed",
				"vulType": []interface{}{
					"We need to quantify the auxiliary PNG card!",
					"reboot",
					"port",
				},
				"cveId":       "CVE-0xc545",
				"description": "The SMTP transmitter is down, transmit the optical application so we can parse the PCI feed!",
			})
		})
		convey.Convey("sort_search_by_term_query", func() {
			var res map[string]interface{}
			docID, err := s.esClient.SearchDocByTermQuery(context.Background(), "analysis-assets_poc",
				map[string]interface{}{}, &res, SearchWithSort("scanAt", false))
			convey.So(err, convey.ShouldBeNil)
			convey.So(docID, convey.ShouldEqual, "")
			convey.So(res, convey.ShouldResemble, map[string]interface{}{})
		})
	})
}
func (s *MainSuite) Test_SearchDocByTermQuery() {
	convey.SkipConvey("Test_SearchDocByTermQuery", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainSuite", "Test_SearchDocByTermQuery")
		})
		convey.Convey("normal_search_by_term_query", func() {
			var res map[string]interface{}
			docID, err := s.esClient.SearchDocByTermQuery(s.ctx, "analysis-assets_poc",
				map[string]interface{}{"target": "directgenerate.name", "fileName": "wordpress_wp_license_file_read.json"}, &res)
			convey.So(err, convey.ShouldBeNil)
			convey.So(docID, convey.ShouldEqual, "zv1h-YMBRzVnG-Ha1e8n")
			convey.So(res, convey.ShouldResemble, map[string]interface{}{
				"target":       "directgenerate.name",
				"scanAt":       "2022-10-21 15:11:20",
				"fileName":     "wordpress_wp_license_file_read.json",
				"hostInfo":     "http://124.236.103.51",
				"name":         "wordpress内容管理系统后门文件wp-license.php页面file参数任意文件读取",
				"vulnerable":   true,
				"data":         "",
				"lastResponse": "",
				"level":        "1",
				"method":       "DELETE",
				"path":         "https://www.districtmesh.name/virtual/sticky/technologies/sexy",
				"vulURL":       "http://www.legacyorchestrate.biz/synergistic/maximize/infrastructures/best-of-breed",
				"vulType": []interface{}{
					"We need to quantify the auxiliary PNG card!",
					"reboot",
					"port",
				},
				"cveId":       "CVE-0xc545",
				"description": "The SMTP transmitter is down, transmit the optical application so we can parse the PCI feed!",
			})
		})
		convey.Convey("sort_search_by_term_query", func() {
			var res map[string]interface{}
			docID, err := s.esClient.SearchDocByTermQuery(context.Background(), "analysis-assets_poc",
				map[string]interface{}{}, &res, SearchWithSort("scanAt", false))
			convey.So(err, convey.ShouldBeNil)
			convey.So(docID, convey.ShouldEqual, "")
			convey.So(res, convey.ShouldResemble, map[string]interface{}{})
		})
	})
}
