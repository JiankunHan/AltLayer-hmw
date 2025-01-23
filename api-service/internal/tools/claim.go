package claim

import (
	"fmt"
	"log"
	"math/big"
	"strconv"

	ganache_connector "hw-app/internal/middleware"
	mysql_connector "hw-app/internal/repository"

	"github.com/gin-gonic/gin"
)

type tokenClaimRes struct {
	User                 string `json:"user"`
	LastInsertID         int64  `json:"last_inesrt_id"`
	TransactionCompleted bool   `json:"transaction_completed"`
	TransactionHash      string `json:"transaction_hash"`
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

func CreateClaimReq(c *gin.Context) {
	user := c.Query("user")
	transactionAmount := c.Query("amount")
	operation := c.Query("operation")
	address := c.Query("contract_address")
	privateKey := c.Query("private_key")
	var claimType uint8
	if operation == "withdraw" {
		claimType = 1
	} else {
		claimType = 0
	}
	var res tokenClaimRes

	if !AuthStaff(user) {
		c.JSON(401, gin.H{"error": "User unauthorized"})
		return
	}

	treasuryAdequate, err := TreasuryAdequate(transactionAmount, claimType, address, privateKey)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if treasuryAdequate == false {
		err := fmt.Errorf("Withdraw claim cannot create due to balance deficiency")
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	lastInsertID, err := mysql_connector.CreateClaimReq(user, claimType, transactionAmount, address, privateKey)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	log.Println("Create a token claim record, last insert ID: ", lastInsertID)
	res.User = user
	res.LastInsertID = lastInsertID
	res.TransactionCompleted = false
	res.TransactionHash = ""
	c.JSON(200, res)
}

func GetTokenClaims(c *gin.Context) {
	user := c.Query("user")
	id_str := c.Query("id")
	status_str := c.Query("status")
	type_str := c.Query("type")

	var user_pt *string
	var id_pt *int
	var status_pt *int
	var type_pt *int
	if user != "" {
		user_pt = &user
	}
	if id_str != "" {
		id, err := strconv.Atoi(id_str)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		id_pt = &id
	}
	if status_str != "" {
		status, err := strconv.Atoi(status_str)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		status_pt = &status
	}
	if type_str == "withdraw" {
		claimType := 1
		type_pt = &claimType
	} else if type_str == "deposit" {
		claimType := 0
		type_pt = &claimType
	}

	claims, err := mysql_connector.GetTokenClaim(user_pt, id_pt, status_pt, type_pt)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, claims)
}
