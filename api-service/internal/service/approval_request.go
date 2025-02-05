package service

import (
	"fmt"
	domain "hw-app/internal/domain"
	utils "hw-app/internal/utils"
	"net/http"
	"strconv"
	"time"
)

// process http request, and enqueue tasks
func HandleApprovalRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	operation := query.Get("operation") //POST request, can be approval/unapproval
	user := query.Get("user")
	operationType := query.Get("type") // GET request, type can be approval/unapproval
	claimID_str := query.Get("claim_id")
	approvalID_str := query.Get("approval_id")
	approvalStatus_str := query.Get("approval_status")
	var task domain.Task

	//approvalStatus_str, approvalID and claimID_str can be empty string when http url does not contain this parameter
	claimID, err := strconv.ParseInt(claimID_str, 10, 64)
	if err != nil && len(claimID_str) != 0 {
		utils.HttpResponse(w, "Wrong format in parameter - claim_id", http.StatusBadRequest)
		return
	}
	if len(claimID_str) == 0 {
		claimID = -1
	}
	approvalID, err := strconv.ParseInt(approvalID_str, 10, 64)
	//approvalStatus_str and claimID_str can be empty string when http url does not contain this parameter
	if err != nil && len(approvalID_str) != 0 {
		utils.HttpResponse(w, "Wrong format in parameter - approval_id", http.StatusBadRequest)
		return
	}
	if len(approvalID_str) == 0 {
		approvalID = -1
	}
	approvalStatus, err := strconv.Atoi(approvalStatus_str)
	if err != nil && len(approvalStatus_str) != 0 {
		utils.HttpResponse(w, "Wrong format in parameter - claim_status", http.StatusBadRequest)
		return
	}
	if len(approvalStatus_str) == 0 {
		approvalStatus = -1
	}

	taskId := utils.GenerateTaskID()
	w.Header().Set("Content-Type", "application/json")
	// 为每个请求创建一个唯一的 response 通道
	respChan := make(chan string, 1)
	// 存储请求的 response 通道
	utils.ResponseMap.Store(taskId, respChan)

	taskInfo := domain.TaskInfo{
		Type:           operationType,
		User:           user,
		ClaimId:        claimID,
		ApprovalId:     approvalID,
		ApprovalStatus: approvalStatus,
		Operation:      operation,
	}

	reqMethod := ""
	switch r.Method {
	case "GET":
		reqMethod = "GET"
	case "POST":
		reqMethod = "POST"
	default:
		// return '405 Method Not Allowed'
		utils.HttpResponse(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	taskTatus := 0
	task = domain.Task{
		ID:        taskId,
		TaskInfo:  taskInfo,
		Status:    taskTatus,
		ReqMethod: reqMethod,
		TaskType:  1,
		RespChan:  respChan,
	}

	select {
	case utils.TaskQueue <- task:
		// in queue, when task queue is not full
		fmt.Println("task enqueue:", task.ID)
	default:
		// return 503 if task queue is full
		utils.HttpResponse(w, "Task queue is full, try again later", http.StatusServiceUnavailable)
	}

	// 等待 worker 处理完成，设置超时控制
	select {
	case result := <-respChan:
		fmt.Fprintf(w, "%s", result)
	case <-time.After(5 * time.Second): // 超时返回
		http.Error(w, "Request timeout", http.StatusGatewayTimeout)
	}
	// 清理 map，防止内存泄漏
	utils.ResponseMap.Delete(taskId)
}
