package domain

import "time"

//define table fields in Mysql

//TokenCliams table
type Claim struct {
	ID              int64     `json:"id"`
	Claimer         string    `json:"claimer"`
	ContractAddress string    `json:"contract_address"`
	PrivateKey      string    `json:"private_key"`
	ClaimType       int8      `json:"claim_type"`
	Amount          string    `json:"amount"`
	ClaimStatus     int8      `json:"claim_status"`
	TransactionHash string    `json:"transaction_hash"`
	CreatedTime     time.Time `json:"created_time"`
	UpdatedTime     time.Time `json:"updated_time"`
}

//WithdrawApprovals table
type Approval struct {
	ID            int64     `json:"id"`
	ClaimId       int64     `json:"claim_id"`
	Approver      string    `json:"approver"`
	ApproveStatus int8      `json:"approve_status"`
	CreatedTime   time.Time `json:"created_time"`
	UpdatedTime   time.Time `json:"updated_time"`
}
