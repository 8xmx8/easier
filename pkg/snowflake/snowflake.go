package snowflake

import (
	"crypto/sha256"
	"encoding/binary"
	"github.com/bwmarrin/snowflake"
)

type Snowflake struct {
	nodeStr string
	node    *snowflake.Node
}

func NewSnowflake(nodeStr string) *Snowflake {
	// nodeStr 转为唯一int64
	nodeID := generateNodeID(nodeStr) % 1024
	node, err := snowflake.NewNode(nodeID)
	if err != nil {
		panic(err)
	}
	return &Snowflake{
		nodeStr: nodeStr,
		node:    node,
	}
}
func (s *Snowflake) NextId() int64 {
	return s.node.Generate().Int64()
}

// 生成节点ID的核心逻辑
func generateNodeID(nodeStr string) int64 {
	h := sha256.New()
	h.Write([]byte(nodeStr))
	hashBytes := h.Sum(nil)
	// 使用前8字节与后8字节进行异或
	first8 := binary.BigEndian.Uint64(hashBytes[:8])
	last8 := binary.BigEndian.Uint64(hashBytes[24:32])
	return int64(first8^last8) & 0x7FFFFFFFFFFFFFFF
}
