package domain

import "net/http"

//define struct fields in the programme

// task struct
type Task struct {
	ID        int      `json:"id"`
	TaskInfo  TaskInfo `json:"task_info"`
	Status    int      `json:"status"`
	ReqMethod string   `json:"req_method"`
	Response  http.ResponseWriter
}

//status: 0 - pending, 1 - working, 2 - completed, 3 - retrying, 4 - failed

type TaskInfo struct {
	Type         string `json:"request_type"`
	Amount       string `json:"amount"`
	User         string `json:"user"`
	CliamId      int    `json:"cliam_id"`
	ClaimStatus  int8   `json:"claim_status"`
	ContractAddr string `json:"contract_address"`
	PrivateKey   string `json:"private_key"`
	Operation    string `json:"operation"`
}

type Result struct {
	TaskID   int    `json:"task_id"`
	Message  string `json:"message"`
	Status   string `json:"status"`
	Response http.ResponseWriter
}

//status: http status

type Transaction struct {
	TaskID       int    `json:"task_id"`
	ContractAddr string `json:"contract_addr"`
	Amount       string `json:"amount"`
	PrivateKey   string `json:"private_key"`
	TrxHash      string `json:"trx_id"`
	Status       string `json:"status"`
}
