package domain

//define struct fields in the programme

// task struct
type Task struct {
	ID        int64       `json:"id"`
	TaskInfo  TaskInfo    `json:"task_info"`
	Status    int         `json:"status"`
	ReqMethod string      `json:"req_method"`
	TaskType  int8        `json:"task_type"` //0: token claim; 1: approval
	RespChan  chan string // 用于 worker 回传结果
}

//status: 0 - pending, 1 - working, 2 - completed, 3 - retrying, 4 - failed

type TaskInfo struct {
	Type           string `json:"operation_type"` //withdraw/deposit for tokenClaim api, approval/unapproval for approval api
	Amount         string `json:"amount"`
	User           string `json:"user"`
	CliamId        int    `json:"cliam_id"`
	ClaimStatus    int    `json:"claim_status"`
	ApprovalId     int    `json:"approval_id"`
	ApprovalStatus int    `json:"approval_status"`
	ContractAddr   string `json:"contract_address"`
	PrivateKey     string `json:"private_key"`
	Operation      string `json:"operation"`
}

//status: http status

type Transaction struct {
	TaskID       int64  `json:"task_id"`
	ContractAddr string `json:"contract_addr"`
	Amount       string `json:"amount"`
	PrivateKey   string `json:"private_key"`
	TrxHash      string `json:"trx_id"`
	Status       string `json:"status"`
}

type HttpRequestRes struct {
	User                 string `json:"user"`
	LastInsertID         int64  `json:"last_inesrt_id"`
	TransactionCompleted bool   `json:"transaction_completed"`
	TransactionHash      string `json:"transaction_hash"`
}
