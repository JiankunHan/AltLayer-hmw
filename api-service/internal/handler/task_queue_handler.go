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
	switch task.ReqMethod {
	case "GET":
		GetTokenClaims(DB, task, task.TaskInfo.User, task.TaskInfo.CliamId, task.TaskInfo.ClaimStatus, task.TaskInfo.Type)
	case "POST":
		CreateClaimReq(DB, task, task.TaskInfo.User, task.TaskInfo.Amount, task.TaskInfo.Operation, task.TaskInfo.ContractAddr, task.TaskInfo.PrivateKey)
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
		BuildResultAndEnqueue([]byte(fmt.Sprintf("User unauthorized")), http.StatusUnauthorized, task.ID, task)
		return
	}

	treasuryAdequate, err := TreasuryAdequate(transactionAmount, claimType, contractAddr, privateKey)
	if err != nil {
		BuildResultAndEnqueue([]byte(err.Error()), http.StatusInternalServerError, task.ID, task)
		return
	}
	if treasuryAdequate == false {
		err := fmt.Errorf("Withdraw claim cannot create due to balance deficiency")
		BuildResultAndEnqueue([]byte(err.Error()), http.StatusInternalServerError, task.ID, task)
		return
	}

	lastInsertID, err := mysql_connector.CreateClaimReq(DB, user, claimType, transactionAmount, contractAddr, privateKey)
	if err != nil {
		BuildResultAndEnqueue([]byte(err.Error()), http.StatusInternalServerError, task.ID, task)
		return
	}
	log.Println("Create a token claim record, last insert ID: ", lastInsertID)
	res.User = user
	res.LastInsertID = lastInsertID
	res.TransactionCompleted = false
	res.TransactionHash = ""
	resJsonStr, err := json.Marshal(res)
	if err != nil {
		BuildResultAndEnqueue([]byte(err.Error()), http.StatusInternalServerError, task.ID, task)
		return
	}
	fmt.Println(string(resJsonStr))
	BuildResultAndEnqueue(resJsonStr, http.StatusOK, task.ID, task)
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
		BuildResultAndEnqueue([]byte(err.Error()), http.StatusInternalServerError, task.ID, task)
		return
	}
	//transfer from array to string
	claimsJsonStr, err := json.Marshal(claims)
	if err != nil {
		fmt.Println(err.Error())
		BuildResultAndEnqueue([]byte(err.Error()), http.StatusInternalServerError, task.ID, task)
		return
	}
	fmt.Println(string(claimsJsonStr))
	BuildResultAndEnqueue(claimsJsonStr, http.StatusOK, task.ID, task)
}
