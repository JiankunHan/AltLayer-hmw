package utils

import (
	"sync"
)

var taskIDCounter int64
var mu sync.Mutex

// 获取唯一的递增 task ID（时间戳 + 递增计数器）
func GenerateTaskID() int64 {
	mu.Lock()
	defer mu.Unlock()

	taskIDCounter++
	return taskIDCounter
}
