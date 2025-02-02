package handler

import (
	"fmt"
	main "hw-app"
)

// This thread returns result of a request
func ResultHandler() {
	for result := range main.ResultQueue {
		// 模拟处理结果（这里可以是将结果保存到数据库、发送给客户端等）
		// 这里简单地打印结果
		fmt.Printf("Result processed for task %d: %s\n", result.TaskID, result.Status)
	}
}
