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
	// get response channel
	if ch, ok := ResponseMap.Load(taskId); ok {
		responseChan := ch.(chan string)
		responseChan <- res
	}
}
