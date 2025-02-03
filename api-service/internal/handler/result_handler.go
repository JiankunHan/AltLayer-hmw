package handler

import (
	"fmt"
	domain "hw-app/internal/domain"
	utils "hw-app/internal/utils"
	"net/http"
	"sync"
)

func HttpResponse(w http.ResponseWriter, res []byte, code int) {
	fmt.Fprintf(w, string(res))
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(code)
	// w.Write(res)
}

func BuildResultAndEnqueue(res []byte, code int, taskId int64, task domain.Task) {
	// result := domain.Result{
	// 	TaskID:     taskId,
	// 	Message:    res,
	// 	StatusCode: code,
	// 	TaskInfo:   task.TaskInfo,
	// }
	// utils.ResultQueue <- result
	// 获取对应的 response 通道
	fmt.Printf("Address of flusher:", taskId)
	if ch, ok := utils.ResponseMap.Load(taskId); ok {
		responseChan := ch.(chan string)
		responseChan <- string(res)
	}
	/*select {
	case ResultQueue <- result:
		// in queue, when task queue is not full
	default:
		// return directly if task queue is full
		HttpResponse(w, res, code)
	}*/
}

// This thread returns result of a request
func ResultHandler(wg *sync.WaitGroup) {
	defer wg.Done() // 完成任务后减少 WaitGroup 的计数
	for result := range utils.ResultQueue {
		fmt.Printf("Result processed for task %d: %d %s\n", result.TaskID, result.StatusCode, string(result.Message))
		// w := result.Response
		// w.Header().Set("Content-Type", "application/json")
		// result.Response.WriteHeader(result.StatusCode)
		// w.Write(result.Message)
		// fmt.Fprintf(result.Response, string(result.Message))
		// fmt.Printf("Address of flusher: %p\n", &result.Flusher)
		// result.Flusher.Flush()
	}
}
