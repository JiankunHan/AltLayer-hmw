package utils

import (
	"sync"
)

var taskIDCounter int64
var mu sync.Mutex

// generate unique taskID
func GenerateTaskID() int64 {
	mu.Lock()
	defer mu.Unlock()

	taskIDCounter++
	return taskIDCounter
}
