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
	"net/http"
	"sync"
	"time"
)

// This thread interacts with Ganache
func GanacheHandler(DB *sql.DB, maxRetryTimes int, wg *sync.WaitGroup) {
	defer wg.Done() // WaitGroup minus 1 after finish processing in this thread
	for transaction := range utils.TransactionQueue {
		log.Printf("Transaction processed for task %d: %s\n", transaction.TaskID, transaction.Status)
		processTransaction(DB, transaction, maxRetryTimes)
	}
}

func processTransaction(DB *sql.DB, transaction domain.Transaction, maxRetryTimes int) {
	claimType := transaction.ClaimType
	contractAddr := transaction.ContractAddr
	amount := transaction.Amount
	privateKey := transaction.PrivateKey
	claimID := transaction.ClaimID
	transaction.Status = 1
	var res domain.HttpRequestRes
	//raise a transaction based on claim
	if claimType == 0 {
		trxhash, err := execDepositTransaction(contractAddr, amount, privateKey)
		if err != nil {
			transaction.Status = 3
			if maxRetryTimes == int(transaction.RetryTimes) {
				err := fmt.Errorf("Transaction failed after retrying %d times for claim: %d", transaction.RetryTimes, claimID)
				transaction.Task.Status = 4
				utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, transaction.TaskID, transaction.Task)
			}
			transaction.RetryTimes++
			time.Sleep(time.Second * 2)
			utils.TransactionQueue <- transaction
			return
		}
		err = mysql_connector.UpdateTokenClaimsStatus(DB, 1, claimID, trxhash)
		if err != nil {
			err := fmt.Errorf("Transaction completed but failed to update database for claim id: %d, manual interference required", claimID)
			utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, transaction.TaskID, transaction.Task)
			return
		}
		res.TransactionCompleted = true
		res.TransactionHash = trxhash
		res.User = transaction.Task.TaskInfo.User
		resJsonStr, err := json.Marshal(res)
		if err != nil {
			utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, transaction.Task.ID, transaction.Task)
			return
		}
		fmt.Println(string(resJsonStr))
		utils.BuildResultAndEnqueue(string(resJsonStr), http.StatusOK, transaction.Task.ID, transaction.Task)
	} else if claimType == 1 {
		trxhash, err := execWithdrawTransaction(contractAddr, amount, privateKey)
		if err != nil {
			transaction.Status = 3
			if maxRetryTimes == int(transaction.RetryTimes) {
				err := fmt.Errorf("Transaction failed after retrying %d times for claim: %d", transaction.RetryTimes, claimID)
				transaction.Task.Status = 4
				utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, transaction.TaskID, transaction.Task)
			}
			transaction.RetryTimes++
			time.Sleep(time.Second * 2)
			utils.TransactionQueue <- transaction
			return
		}
		err = mysql_connector.UpdateTokenClaimsStatus(DB, 1, claimID, trxhash)
		if err != nil {
			err := fmt.Errorf("Transaction completed but failed to update database for claim id: %d, manual interference required", claimID)
			utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, transaction.TaskID, transaction.Task)
			return
		}
		res.TransactionCompleted = true
		res.TransactionHash = trxhash
		res.User = transaction.Task.TaskInfo.User
		resJsonStr, err := json.Marshal(res)
		if err != nil {
			utils.BuildResultAndEnqueue(err.Error(), http.StatusInternalServerError, transaction.Task.ID, transaction.Task)
			return
		}
		fmt.Println(string(resJsonStr))
		utils.BuildResultAndEnqueue(string(resJsonStr), http.StatusOK, transaction.Task.ID, transaction.Task)
	}
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
