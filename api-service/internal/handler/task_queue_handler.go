package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	domain "hw-app/internal/domain"
	ganache_connector "hw-app/internal/middleware"
	mysql_connector "hw-app/internal/repository"
	utils "hw-app/internal/utils"
	"log"
	"math/big"
	"net/http"
	"sync"
)

// DB connections for each RequestHandler thread
var PoolDB []*sql.DB

// RequestHandler, deal with tasks in the queue
func RequestHandler(id int, wg *sync.WaitGroup) {
	defer wg.Done() // 完成任务后减少 WaitGroup 的计数

	for task := range utils.TaskQueue {
		fmt.Printf("req_handler %d is processing task %d\n", id, task.ID)
		//RequestHandler entrance function
		dispatchHttpRequest(task, PoolDB[id])
	}
}

func dispatchHttpRequest(task domain.Task, DB *sql.DB) {
	if task.TaskType == 0 {
		switch task.ReqMethod {
		case "GET":
			GetTokenClaims(DB, task, task.TaskInfo.User, task.TaskInfo.CliamId, task.TaskInfo.ClaimStatus, task.TaskInfo.Type)
		case "POST":
			CreateClaimReq(DB, task, task.TaskInfo.User, task.TaskInfo.Amount, task.TaskInfo.Operation, task.TaskInfo.ContractAddr, task.TaskInfo.PrivateKey)
		}
	} else if task.TaskType == 1 {
		switch task.ReqMethod {
		case "GET":
			GetClaimApproval(DB, task, task.TaskInfo.User, task.TaskInfo.CliamId, task.TaskInfo.ApprovalId, task.TaskInfo.ApprovalStatus)
		case "POST":
			CreateClaimApproval(DB, task, task.TaskInfo.User, task.TaskInfo.Operation, task.TaskInfo.CliamId)
		}
	}

}

func TreasuryAdequate(transactionAmount string, claimType uint8, address string, privateKey string) (bool, error) {
	if claimType == 0 {
		return true, nil
	}
	bigTrxAmount := new(big.Int)
	bigTrxAmount, success := bigTrxAmount.SetString(transactionAmount, 10)
	if !success {
		err := fmt.Errorf("Invalid token amount: %s", transactionAmount)
		return false, err
	}
	balance, err := ganache_connector.GetBalance(address, privateKey)
	if err != nil {
		return false, err
	}
	fmt.Println("Balance:", balance)
	fmt.Println("transactionAmount:", bigTrxAmount)
	treasuryAdequate := balance.Cmp(bigTrxAmount) >= 0
	return treasuryAdequate, nil
}

func CreateClaimReq(DB *sql.DB, task domain.Task, user string, transactionAmount string, operation string, contractAddr string, privateKey string) {
	var claimType uint8
	if operation == "withdraw" {
		claimType = 1
	} else {
		claimType = 0
	}
	var res domain.HttpRequestRes

	if !utils.AuthStaff(user) {
		utils.BuildResultAndEnqueue(fmt.Sprintf("User unauthorized"), http.StatusUnauthorized, task.ID, task)
		return
	}

	treasuryAdequate, err := TreasuryAdequate(transactionAmount, claimType, contractAddr, privateKey)
	if err != nil {
		utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, task.ID, task)
		return
	}
	if treasuryAdequate == false {
		err := fmt.Errorf("Withdraw claim cannot create due to balance deficiency")
		utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, task.ID, task)
		return
	}

	lastInsertID, err := mysql_connector.CreateClaimReq(DB, user, claimType, transactionAmount, contractAddr, privateKey)
	if err != nil {
		utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, task.ID, task)
		return
	}
	log.Println("Create a token claim record, last insert ID: ", lastInsertID)
	res.User = user
	res.LastInsertID = lastInsertID
	res.TransactionCompleted = false
	res.TransactionHash = ""
	resJsonStr, err := json.Marshal(res)
	if err != nil {
		utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, task.ID, task)
		return
	}
	fmt.Println(string(resJsonStr))
	utils.BuildResultAndEnqueue(string(resJsonStr), http.StatusOK, task.ID, task)
}

func GetTokenClaims(DB *sql.DB, task domain.Task, user string, id int, status int, claimTypeStr string) {
	var user_pt *string
	var id_pt *int
	var status_pt *int
	var type_pt *int
	if user != "" {
		user_pt = &user
	}
	id_pt = &id
	status_pt = &status
	if claimTypeStr == "withdraw" {
		claimType := 1
		type_pt = &claimType
	} else if claimTypeStr == "deposit" {
		claimType := 0
		type_pt = &claimType
	}

	claims, err := mysql_connector.GetTokenClaim(DB, user_pt, id_pt, status_pt, type_pt)
	if err != nil {
		fmt.Println(err.Error())
		utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, task.ID, task)
		return
	}
	//transfer from array to string
	claimsJsonStr, err := json.Marshal(claims)
	if err != nil {
		fmt.Println(err.Error())
		utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, task.ID, task)
		return
	}
	fmt.Println(string(claimsJsonStr))
	utils.BuildResultAndEnqueue(string(claimsJsonStr), http.StatusOK, task.ID, task)
}

func checkIdenticalApproval(DB *sql.DB, user *string, claimID *int, approve_status *int) (bool, error) {
	//if there is a record in WithdrawApprovals with same approver, claim_id and approve_status, reject this request
	claims, err := mysql_connector.GetApprovals(DB, user, nil, claimID, approve_status)
	if err != nil {
		return true, err
	}
	if len(claims) > 0 {
		return true, nil
	}
	return false, nil
}

func updateIfRecordExist(DB *sql.DB, user *string, claimID *int, approve_status int) (int64, error) {
	//if there is a record in WithdrawApprovals with same approver, claim_id, update the approve_status
	claims, err := mysql_connector.GetApprovals(DB, user, nil, claimID, nil)
	if err != nil {
		return 0, err
	}
	if len(claims) > 0 {
		err = mysql_connector.UpdateWithdrawApprovalsStatus(DB, approve_status, int(claims[0].ID))
		if err != nil {
			return 0, err
		}
		return int64(claims[0].ID), nil
	}
	return 0, nil
}

func CreateClaimApproval(DB *sql.DB, task domain.Task, user string, operation string, claimID int) {
	var approve_status int
	var res domain.HttpRequestRes

	if operation == "approve" {
		approve_status = 1
	} else {
		approve_status = 0
	}

	if !utils.AuthManager(user) {
		utils.BuildResultAndEnqueue("Manager unauthorized to approve/unapprove", http.StatusInternalServerError, task.ID, task)
		return
	}

	hasIdenticalApproval, err := checkIdenticalApproval(DB, &user, &claimID, &approve_status)
	if err != nil {
		fmt.Println(err.Error())
		utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, task.ID, task)
		return
	}
	if hasIdenticalApproval {
		utils.BuildResultAndEnqueue("Cannot operate approval/unapproval again", http.StatusInternalServerError, task.ID, task)
		return
	}

	updatedApprovalRecord, err := updateIfRecordExist(DB, &user, &claimID, approve_status)
	if err != nil {
		fmt.Println(err.Error())
		utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, task.ID, task)
		return
	}
	if updatedApprovalRecord != 0 {
		transactionCompleted, trxHash, err := CheckAndRaiseTokenTransaction(DB, claimID)
		if err != nil {
			fmt.Println(err.Error())
			utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, task.ID, task)
			return
		}
		res.User = user
		res.LastInsertID = updatedApprovalRecord
		res.TransactionCompleted = transactionCompleted
		res.TransactionHash = trxHash
		resJsonStr, err := json.Marshal(res)
		if err != nil {
			utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, task.ID, task)
			return
		}
		fmt.Println(string(resJsonStr))
		utils.BuildResultAndEnqueue(string(resJsonStr), http.StatusOK, task.ID, task)
		return
	}

	lastInsertID, err := mysql_connector.CreateClaimApproval(DB, user, claimID, approve_status)
	if err != nil {
		fmt.Println(err.Error())
		utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, task.ID, task)
		return
	}
	log.Println("Create a token claim record, last insert ID: ", lastInsertID)
	transactionCompleted, trxHash, err := CheckAndRaiseTokenTransaction(DB, claimID)
	if err != nil {
		fmt.Println(err.Error())
		utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, task.ID, task)
		return
	}
	res.User = user
	res.LastInsertID = lastInsertID
	res.TransactionCompleted = transactionCompleted
	res.TransactionHash = trxHash
	resJsonStr, err := json.Marshal(res)
	if err != nil {
		utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, task.ID, task)
		return
	}
	fmt.Println(string(resJsonStr))
	utils.BuildResultAndEnqueue(string(resJsonStr), http.StatusOK, task.ID, task)
}

func GetClaimApproval(DB *sql.DB, task domain.Task, approver string, claimID int, approvalID int, status int) {
	var approver_pt *string
	var id_pt *int
	var claim_id_pt *int
	var status_pt *int

	if approver != "" {
		approver_pt = &approver
	}
	id_pt = &approvalID
	claim_id_pt = &claimID
	status_pt = &status

	approvals, err := mysql_connector.GetApprovals(DB, approver_pt, id_pt, claim_id_pt, status_pt)
	if err != nil {
		fmt.Println(err.Error())
		utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, task.ID, task)
		return
	}
	//transfer from array to string
	approvalsJsonStr, err := json.Marshal(approvals)
	if err != nil {
		fmt.Println(err.Error())
		utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, task.ID, task)
		return
	}
	fmt.Println(string(approvalsJsonStr))
	utils.BuildResultAndEnqueue(string(approvalsJsonStr), http.StatusOK, task.ID, task)
}

func isApprovalsAdequate(DB *sql.DB, claimID int) (bool, error) {
	//check if there are two records in table WithdrawApprovals, approve_status = 1 and claim_id = claimID
	approve_status := 1
	claims, err := mysql_connector.GetApprovals(DB, nil, nil, &claimID, &approve_status)
	if err != nil {
		return false, err
	}
	if len(claims) >= 2 {
		return true, nil
	}
	return false, nil
}

func execDepositTransaction(contractAddr string, amount string, privateKey string) (string, error) {
	trxhash, err := ganache_connector.DepositTransaction(contractAddr, amount, privateKey)
	if err != nil {
		return "", err
	}
	return trxhash, nil
}

func execWithdrawTransaction(contractAddr string, amount string, privateKey string) (string, error) {
	trxhash, err := ganache_connector.WithDrawTransaction(contractAddr, amount, privateKey)
	if err != nil {
		return "", err
	}
	return trxhash, nil
}

func CheckAndRaiseTokenTransaction(DB *sql.DB, claimID int) (bool, string, error) {
	readyRaiseTrans, err := isApprovalsAdequate(DB, claimID)
	if err != nil {
		return false, "", err
	}
	if readyRaiseTrans == false {
		return false, "", nil
	}

	//get original claim
	claims, err := mysql_connector.GetTokenClaim(DB, nil, &claimID, nil, nil)
	if err != nil {
		return false, "", err
	}
	if len(claims) != 1 {
		err := fmt.Errorf("Data misaligned for claim id: %d", claimID)
		return false, "", err
	}
	if claims[0].ClaimStatus != 0 {
		//transaction completed and cliam been closed
		return true, "", nil
	}
	contractAddr := claims[0].ContractAddress
	amount := claims[0].Amount
	claimType := claims[0].ClaimType
	privateKey := claims[0].PrivateKey
	var trxhash string

	//raise a transaction based on claim
	if claimType == 0 {
		trxhash, err = execDepositTransaction(contractAddr, amount, privateKey)
		if err != nil {
			err := fmt.Errorf("Transaction failed for claim id: %d, manual interference required", claimID)
			return false, "", err
		}
		err := mysql_connector.UpdateTokenClaimsStatus(DB, 1, claimID, trxhash)
		if err != nil {
			err := fmt.Errorf("Transaction completed but failed to update database for claim id: %d, manual interference required", claimID)
			return false, "", err
		}
	} else if claimType == 1 {
		trxhash, err = execWithdrawTransaction(contractAddr, amount, privateKey)
		if err != nil {
			err := fmt.Errorf("Transaction failed for claim id: %d, manual interference required", claimID)
			return false, "", err
		}
		err := mysql_connector.UpdateTokenClaimsStatus(DB, 1, claimID, trxhash)
		if err != nil {
			err := fmt.Errorf("Transaction completed but failed to update database for claim id: %d, manual interference required", claimID)
			return false, "", err
		}
	}

	return true, trxhash, nil
}
