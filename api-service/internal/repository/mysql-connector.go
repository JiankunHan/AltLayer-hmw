package mysql_connector

import (
	"log"
	"strconv"
	"time"

	domain "hw-app/internal/domain"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// Global DB variable
// var DB *sql.DB

func IntializeDBConn() (*sql.DB, error) {
	// initalize Mysql connection. It can be a connection pool. But we use one connection for simplification.
	DB, err := InitDB()
	if err != nil {
		log.Fatal("Fail to create database connection: ", err)
		return nil, err
	}
	// defer CloseDB(DB)
	return DB, nil
}

// 初始化 MySQL 连接
func InitDB() (*sql.DB, error) {
	dsn := "root:Jiankun9598+@tcp(mysql:3306)/homework_db?charset=utf8mb4&parseTime=True&loc=Local"
	DB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Check if the connection is valid
	if err := DB.Ping(); err != nil {
		return nil, err
	}

	log.Println("Database connected successfully.")
	return DB, nil
}

// CloseDB closes the database connection
func CloseDB(DB *sql.DB) error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

func CreateClaimReq(DB *sql.DB, user string, claimType uint8, transactionAmount string, address string, privateKey string) (int64, error) {
	if DB == nil {
		log.Fatal("Database connection is lost")
	}

	insertQuery := "INSERT INTO TokenClaims (claimer, contract_address, private_key, claim_type, amount, claim_status, created_time, updated_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	currentTime := time.Now()
	result, err := DB.Exec(insertQuery, user, address, privateKey, claimType, transactionAmount, 0, currentTime, currentTime)

	if err != nil {
		return 0, err
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastInsertID, nil
}

func CreateClaimApproval(DB *sql.DB, user string, claimId int64, status int) (int64, error) {
	if DB == nil {
		log.Fatal("Database connection is lost")
	}

	insertQuery := "INSERT INTO WithdrawApprovals (claim_id, approver, approve_status, created_time, updated_time) VALUES (?, ?, ?, ?, ?)"
	currentTime := time.Now()
	result, err := DB.Exec(insertQuery, claimId, user, status, currentTime, currentTime)

	if err != nil {
		return 0, err
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastInsertID, nil
}

func GetTokenClaim(DB *sql.DB, user *string, claimID *int64, status *int, claimType *int) ([]domain.Claim, error) {
	if DB == nil {
		log.Fatal("Database connection is lost")
	}

	var claims []domain.Claim
	whereConditionExist := false

	query := "SELECT id, claimer, contract_address, private_key, claim_type, amount, claim_status, created_time, updated_time FROM TokenClaims"
	if user != nil || (claimID != nil && *claimID >= 0) || (status != nil && *status >= 0) || claimType != nil {
		query += " where "
	}
	if user != nil {
		userClause := "claimer = '"
		userClause += *user
		userClause += "'"
		query += userClause
		whereConditionExist = true
	}
	if claimID != nil && *claimID >= 0 {
		var idClause string
		claimId := strconv.FormatInt(*claimID, 10)
		if whereConditionExist {
			idClause = " and id = "
			idClause += claimId
		} else {
			idClause = " id = "
			idClause += claimId
		}
		query += idClause
		whereConditionExist = true
	}
	if status != nil && *status >= 0 {
		var statusClause string
		Status := strconv.Itoa(*status)
		if whereConditionExist {
			statusClause = " and claim_status = "
			statusClause += Status
		} else {
			statusClause = " claim_status = "
			statusClause += Status
		}
		query += statusClause
		whereConditionExist = true
	}
	if claimType != nil {
		var typeClause string
		ClaimType := strconv.Itoa(*claimType)
		if whereConditionExist {
			typeClause = " and claim_type = "
			typeClause += ClaimType
		} else {
			typeClause = " claim_type = "
			typeClause += ClaimType
		}
		query += typeClause
	}

	query += ";"
	log.Println(query)

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var claim domain.Claim
		if err := rows.Scan(&claim.ID, &claim.Claimer, &claim.ContractAddress, &claim.PrivateKey, &claim.ClaimType, &claim.Amount, &claim.ClaimStatus, &claim.CreatedTime, &claim.UpdatedTime); err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return claims, nil
}

func UpdateTokenClaimsStatus(DB *sql.DB, claim_status int, id int64, trxhash string) error {
	if DB == nil {
		log.Fatal("Database connection is lost")
	}
	query := "UPDATE TokenClaims SET claim_status = ?, updated_time = ?, transaction_hash = ? WHERE id = ?"
	currentTime := time.Now()
	_, err := DB.Exec(query, claim_status, currentTime, trxhash, id)
	return err
}

func GetApprovals(DB *sql.DB, user *string, ID *int64, claimID *int64, approve_status *int) ([]domain.Approval, error) {
	if DB == nil {
		log.Fatal("Database connection is lost")
	}

	var approvals []domain.Approval
	whereConditionExist := false

	query := "SELECT id, claim_id, approver, approve_status, created_time, updated_time FROM WithdrawApprovals"
	if user != nil || (ID != nil && *ID >= 0) || (claimID != nil && *claimID >= 0) || (approve_status != nil && *approve_status >= 0) {
		query += " where "
	}
	if user != nil {
		userClause := "approver = '"
		userClause += *user
		userClause += "'"
		query += userClause
		whereConditionExist = true
	}
	if ID != nil && *ID >= 0 {
		var idClause string
		ID := strconv.FormatInt(*ID, 10)
		if whereConditionExist {
			idClause = " and id = "
			idClause += ID
		} else {
			idClause = " id = "
			idClause += ID
		}
		query += idClause
		whereConditionExist = true
	}
	if claimID != nil && *claimID >= 0 {
		var idClause string
		claimId := strconv.FormatInt(*claimID, 10)
		if whereConditionExist {
			idClause = " and claim_id = "
			idClause += claimId
		} else {
			idClause = " claim_id = "
			idClause += claimId
		}
		query += idClause
		whereConditionExist = true
	}
	if approve_status != nil && *approve_status >= 0 {
		var statusClause string
		Status := strconv.Itoa(*approve_status)
		if whereConditionExist {
			statusClause = " and approve_status = "
			statusClause += Status
		} else {
			statusClause = " approve_status = "
			statusClause += Status
		}
		query += statusClause
	}

	query += ";"
	log.Println(query)

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var approval domain.Approval
		if err := rows.Scan(&approval.ID, &approval.ClaimId, &approval.Approver, &approval.ApproveStatus, &approval.CreatedTime, &approval.UpdatedTime); err != nil {
			return nil, err
		}
		approvals = append(approvals, approval)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return approvals, nil
}

func UpdateWithdrawApprovalsStatus(DB *sql.DB, approve_status int, id int) error {
	query := "UPDATE WithdrawApprovals SET approve_status = ?, updated_time = ? WHERE id = ?"
	currentTime := time.Now()
	_, err := DB.Exec(query, approve_status, currentTime, id)
	return err
}
