package handler

import (
	"fmt"
	utils "hw-app/internal/utils"
	"sync"
)

// This thread interacts with Ganache
func GanacheHandler(wg *sync.WaitGroup) {
	defer wg.Done() // 完成任务后减少 WaitGroup 的计数
	for result := range utils.TransactionQueue {
		// 模拟处理结果（这里可以是将结果保存到数据库、发送给客户端等）
		// 这里简单地打印结果
		fmt.Printf("Result processed for task %d: %s\n", result.TaskID, result.Status)
	}
}
