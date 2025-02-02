package handler

import (
	"database/sql"
	"fmt"
	main "hw-app"
	"sync"
)

// DB connections for each RequestHandler thread
var PoolDB []*sql.DB

// RequestHandler, deal with tasks in the queue
func RequestHandler(id int, wg *sync.WaitGroup) {
	defer wg.Done() // 完成任务后减少 WaitGroup 的计数

	for task := range main.TaskQueue {
		fmt.Printf("req_handler %d is processing task %d\n", id, task.ID)
		//RequestHandler entrance function
	}
}
