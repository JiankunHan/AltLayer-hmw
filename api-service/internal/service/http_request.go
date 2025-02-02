package service

import (
	"fmt"
	main "hw-app"
	domain "hw-app/internal/domain"
	"net/http"
	"strconv"
)

// process http request, and enqueue tasks
func HandleClaimRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	operation := query.Get("operation")
	ContractAddr := query.Get("contract_address")
	privateKey := query.Get("private_key")
	amount := query.Get("amount")
	user := query.Get("user")
	cliamType := query.Get("type") // GET request, type can be withdraw/deposit
	claimID_str := query.Get("claim_id")
	claimStatus_str := query.Get("claim_status")
	claimID, err := strconv.Atoi(claimID_str)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Wrong format in parameter - claim_id")))
		return
	}
	claimStatus, err := strconv.Atoi(claimStatus_str)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Wrong format in parameter - claim_status")))
		return
	}
	taskInfo := domain.TaskInfo{
		Type:         cliamType,
		Amount:       amount,
		User:         user,
		CliamId:      claimID,
		ClaimStatus:  int8(claimStatus),
		ContractAddr: ContractAddr,
		PrivateKey:   privateKey,
		Operation:    operation,
	}

	reqMethod := ""
	switch r.Method {
	case "GET":
		reqMethod = "GET"
	case "POST":
		reqMethod = "POST"
	default:
		// 如果请求方法是其他类型，可以返回 405 Method Not Allowed 错误
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method Not Allowed"))
	}

	taskTatus := 0
	task := domain.Task{
		ID:        len(main.TaskQueue) + 1,
		TaskInfo:  taskInfo,
		Status:    taskTatus,
		Response:  w,
		ReqMethod: reqMethod,
	}

	select {
	case main.TaskQueue <- task:
		// in queue, when task queue is not full
	default:
		// return 503 if task queue is full
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Task queue is full, try again later"))
	}
}
