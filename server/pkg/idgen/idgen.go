package idgen

import (
	"fmt"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	mu   sync.RWMutex
	node *snowflake.Node
)

// Init 用指定 node 号初始化全局生成器；node 取值范围 [0,1023]。
func Init(nodeID int64) error {
	n, err := snowflake.NewNode(nodeID)
	if err != nil {
		return fmt.Errorf("idgen: init node %d: %w", nodeID, err)
	}
	mu.Lock()
	node = n
	mu.Unlock()
	return nil
}

// NextID 生成下一个唯一 ID。未初始化时 panic（属于启动装配错误，应尽早暴露）。
func NextID() (uint64, error) {
	mu.RLock()
	n := node
	mu.RUnlock()
	if n == nil {
		return 0, fmt.Errorf("idgen: not initialized, call idgen.Init first")
	}
	return uint64(n.Generate().Int64()), nil
}
