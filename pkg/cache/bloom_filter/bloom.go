package bloomfilter

import (
	"bufio"
	"context"
	"os"

	"github.com/bits-and-blooms/bloom/v3"
)

// BloomFilter 定义布隆过滤器
type BloomFilter struct {
	filter *bloom.BloomFilter
}

type BloomOption func(*BloomFilter) error

// LoadFileWithOption 通过文件加载过滤器样本
func LoadFileWithOption(filepath string) func(*BloomFilter) error {
	return func(bf *BloomFilter) error {
		f, err := os.Open(filepath)
		if err != nil {
			return err
		}
		defer f.Close()
		r := bufio.NewReader(f)
		_, err = bf.filter.ReadFrom(r)
		return err
	}
}

// NewBloom 实例化布隆过滤器
func NewBloom(ctx context.Context, n uint, fp float64, ops ...BloomOption) (*BloomFilter, error) {
	filter := bloom.NewWithEstimates(n, fp)
	bf := &BloomFilter{
		filter: filter,
	}
	for _, op := range ops {
		err := op(bf)
		if err != nil {
			return nil, err
		}
	}
	return bf, nil
}

// Add 向过滤器中增加样本
func (bf *BloomFilter) Add(data []byte) {
	bf.filter.Add(data)
}

// ClearAll 清空过滤器中的所有样本
func (bf *BloomFilter) ClearAll() {
	bf.filter.ClearAll()
}

// Test 如果数据位于 BloomFilter 中，则 Test 返回 true，否则返回 false。如果为 true，则结果可能是误报。如果为 false，则数据肯定不在集合中
func (bf *BloomFilter) Test(data []byte) bool {
	return bf.filter.Test(data)
}

// DownloadToFile 将过滤器的样本保存到本地文件
func (bf *BloomFilter) DownloadToFile(ctx context.Context, filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	_, err = bf.filter.WriteTo(w)
	return err
}

// LoadByFile 从文件中加载
// 先清空过滤后所有样本,重新加载
func (bf *BloomFilter) LoadByFile(ctx context.Context, filepath string) error {
	bf.ClearAll()
	// 加载样本文件
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	_, err = bf.filter.ReadFrom(r)
	return err
}
