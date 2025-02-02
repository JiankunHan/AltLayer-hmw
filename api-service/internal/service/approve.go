package service

import (
	"fmt"
	"log"
	"strconv"

	ganache_connector "hw-app/internal/middleware"
	mysql_connector "hw-app/internal/repository"

	"github.com/gin-gonic/gin"
)

func checkIdenticalApproval(user *string, claimID *int, approve_status *int) (bool, error) {
	//if there is a record in WithdrawApprovals with same approver, claim_id and approve_status, reject this request
	claims, err := mysql_connector.GetApprovals(user, nil, claimID, approve_status)
	if err != nil {
		return true, err
	}
	if len(claims) > 0 {
		return true, nil
	}
	return false, nil
}

func updateIfRecordExist(user *string, claimID *int, approve_status int) (int64, error) {
	//if there is a record in WithdrawApprovals with same approver, claim_id, update the approve_status
	claims, err := mysql_connector.GetApprovals(user, nil, claimID, nil)
	if err != nil {
		return 0, err
	}
	if len(claims) > 0 {
		err = mysql_connector.UpdateWithdrawApprovalsStatus(approve_status, int(claims[0].ID))
		if err != nil {
			return 0, err
		}
		return int64(claims[0].ID), nil
	}
	return 0, nil
}

func CreateClaimApproval(c *gin.Context) {
	claimID_str := c.Query("claim_id")
	user := c.Query("user")
	operation := c.Query("operation")
	claimID, err := strconv.Atoi(claimID_str)
	var approve_status int
	var res tokenClaimRes

	if operation == "approve" {
		approve_status = 1
	} else {
		approve_status = 0
	}

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if !AuthManager(user) {
		c.JSON(401, gin.H{"error": "Manager unauthorized to approve/unapprove"})
		return
	}

	hasIdenticalApproval, err := checkIdenticalApproval(&user, &claimID, &approve_status)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if hasIdenticalApproval {
		c.JSON(400, gin.H{"error": "Cannot operate approval/unapproval again"})
		return
	}

	updatedApprovalRecord, err := updateIfRecordExist(&user, &claimID, approve_status)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if updatedApprovalRecord != 0 {
		transactionCompleted, trxHash, err := CheckAndRaiseTokenTransaction(claimID)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		res.User = user
		res.LastInsertID = updatedApprovalRecord
		res.TransactionCompleted = transactionCompleted
		res.TransactionHash = trxHash
		c.JSON(200, res)
		return
	}

	lastInsertID, err := mysql_connector.CreateClaimApproval(user, claimID, approve_status)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	log.Println("Create a token claim record, last insert ID: ", lastInsertID)
	transactionCompleted, trxHash, err := CheckAndRaiseTokenTransaction(claimID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	res.User = user
	res.LastInsertID = lastInsertID
	res.TransactionCompleted = transactionCompleted
	res.TransactionHash = trxHash
	c.JSON(200, res)
}

func GetClaimApproval(c *gin.Context) {
	approver := c.Query("approver")
	id_str := c.Query("id")
	claim_id_str := c.Query("claim_id")
	status_str := c.Query("status")

	var approver_pt *string
	var id_pt *int
	var claim_id_pt *int
	var status_pt *int
	if approver != "" {
		approver_pt = &approver
	}
	if id_str != "" {
		id, err := strconv.Atoi(id_str)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		id_pt = &id
	}
	if claim_id_str != "" {
		claim_id, err := strconv.Atoi(claim_id_str)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		claim_id_pt = &claim_id
	}
	if status_str != "" {
		status, err := strconv.Atoi(status_str)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		status_pt = &status
	}

	claims, err := mysql_connector.GetApprovals(approver_pt, id_pt, claim_id_pt, status_pt)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, claims)
}

func isApprovalsAdequate(claimID int) (bool, error) {
	//check if there are two records in table WithdrawApprovals, approve_status = 1 and claim_id = claimID
	approve_status := 1
	claims, err := mysql_connector.GetApprovals(nil, nil, &claimID, &approve_status)
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

func CheckAndRaiseTokenTransaction(claimID int) (bool, string, error) {
	readyRaiseTrans, err := isApprovalsAdequate(claimID)
	if err != nil {
		return false, "", err
	}
	if readyRaiseTrans == false {
		return false, "", nil
	}

	//get original claim
	claims, err := mysql_connector.GetTokenClaim(nil, &claimID, nil, nil)
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
		err := mysql_connector.UpdateTokenClaimsStatus(1, claimID, trxhash)
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
		err := mysql_connector.UpdateTokenClaimsStatus(1, claimID, trxhash)
		if err != nil {
			err := fmt.Errorf("Transaction completed but failed to update database for claim id: %d, manual interference required", claimID)
			return false, "", err
		}
	}

	return true, trxhash, nil
}
