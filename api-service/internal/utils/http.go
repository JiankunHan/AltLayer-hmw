package utils

import (
	"fmt"
	domain "hw-app/internal/domain"
	"net/http"
)

func HttpResponse(w http.ResponseWriter, res string, code int) {
	fmt.Fprintf(w, "%s", res)
}

func BuildResultAndEnqueue(res string, code int, taskId int64, task domain.Task) {
	// 获取对应的 response 通道
	if ch, ok := ResponseMap.Load(taskId); ok {
		responseChan := ch.(chan string)
		responseChan <- res
	}
}
