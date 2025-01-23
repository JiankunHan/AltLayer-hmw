CREATE DATABASE IF NOT EXISTS homework_db;

USE homework_db;

CREATE TABLE IF NOT EXISTS TokenClaims (
    id INT AUTO_INCREMENT PRIMARY KEY,
    claimer VARCHAR(100) NOT NULL,
    contract_address VARCHAR(100) NOT NULL,
    private_key VARCHAR(100) NOT NULL,
    claim_type TINYINT NOT NULL,
    amount VARCHAR(200) NOT NULL,
    claim_status TINYINT DEFAULT 0,
    transaction_hash VARCHAR(200),
    created_time DATETIME,
    updated_time DATETIME
);

CREATE TABLE IF NOT EXISTS WithdrawApprovals (
    id INT AUTO_INCREMENT PRIMARY KEY,
    claim_id INT NOT NULL,
    approver VARCHAR(100) NOT NULL,
    approve_status TINYINT DEFAULT 1,
    created_time DATETIME,
    updated_time DATETIME
);

--claim_type: 0 - deposit
--claim_type: 1 - withdraw

--claim_status: 0 - pending
--claim_status: 1 - closed

-- approve_status: 0 - unapproved
-- approve_status: 1 - approved