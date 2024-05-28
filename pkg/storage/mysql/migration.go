package orm

import (
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/spf13/cast"
	"gorm.io/gorm"
)

var Migrate = &Migration{
	version: make(map[int]func(db *gorm.DB, version string) error),
}

// MigrationVersion 数据前一版本表
// nolint
type MigrationVersion struct {
	Version   string    `gorm:"primaryKey"`
	ApplyTime time.Time `gorm:"autoCreateTime"`
}

func (MigrationVersion) TableName() string {
	return "sys_migration"
}

// Migration  定义数据库迁移对象
type Migration struct {
	db      *gorm.DB
	version map[int]func(db *gorm.DB, version string) error
	mutex   sync.Mutex
}

// GetDb 获取需要迁移的数据库连接
func (e *Migration) GetDb() *gorm.DB {
	return e.db
}

// SetVersion 设置需要迁移的文件
func (e *Migration) SetVersion(k int, f func(db *gorm.DB, version string) error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.version[k] = f
}

// Migrate 执行迁移文件
func (e *Migration) Migrate() {
	versions := make([]int, 0)
	for k := range e.version {
		versions = append(versions, k)
	}
	if !sort.IntsAreSorted(versions) {
		sort.Ints(versions)
	}
	var err error
	var count int64
	for _, v := range versions {
		// 查迁移表是否迁移过
		err = e.db.Table("sys_migration").Where("version = ?", v).Count(&count).Error
		if err != nil {
			log.Fatalln(err)
		}
		// 对比已迁移文件不在执迁移
		if count > 0 {
			count = 0
			continue
		}
		// 执行迁移文件
		err = (e.version[v])(e.db, strconv.Itoa(v))
		if err != nil {
			log.Fatalln(err)
		}
	}
}

// GetFilename 获取迁移文件里名前的版本号
func GetFilename(s string) int {
	s = filepath.Base(s)      // 获取迁移文件名 1660204704697_initdb.go
	return cast.ToInt(s[:13]) // 获取迁移文件的版本号 1660204704697
}
