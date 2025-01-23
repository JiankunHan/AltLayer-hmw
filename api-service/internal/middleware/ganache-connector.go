package ganache_connector

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// 定义模型
type transactionRes struct {
	Address           string `json:"address"`
	TransactionHash   string `json:"transactionHash"`
	TransactionAmount int64  `json:"transactionAmount"`
}

const (
	// contractAddr = "0x5f8e26fAcC23FA4cbd87b8d9Dbbd33D5047abDE1"
	// privateKey = "0x4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"
	retryAttempts = 5               // 最大重试次数
	retryDelay    = time.Second * 2 // 每次重试的延迟时间
)

func DepositTransaction(contractAddr string, depositAmount string, privateKey string) (string, error) {
	ganacheURL := os.Getenv("GANACHE_URL")
	bigTrxAmount := new(big.Int)
	bigTrxAmount, success := bigTrxAmount.SetString(depositAmount, 10)
	if !success {
		err := fmt.Errorf("Invalid token amount: %s", depositAmount)
		return "", err
	}
	// 创建一个有效的上下文
	ctx := context.Background()

	// 连接到 Ganache 节点
	client, err := ethclient.DialContext(ctx, ganacheURL)
	if err != nil {
		log.Println("Failed to connect to the Ethereum client: %v", err)
		return "", err
	}

	fmt.Println("Successfully connected to Ganache: ", client)

	// 加载私钥
	var privatekey = strings.TrimPrefix(privateKey, "0x") // 移除前缀
	key, err := crypto.HexToECDSA(privatekey)
	if err != nil {
		log.Println("Failed to load private key: %v", err)
		return "", err
	}

	// 创建一个授权者
	fromAddress := crypto.PubkeyToAddress(key.PublicKey)
	fmt.Println("From Address: ", fromAddress.Hex())

	// 获取当前的 Gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Println("Failed to get gas price: %v", err)
		return "", err
	}
	fmt.Println("Gas price: ", gasPrice)

	// 加载合约 ABI
	contractABI, err := abi.JSON(strings.NewReader(`[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"balances","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function","constant":true},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function","constant":true},{"inputs":[],"name":"deposit","outputs":[],"stateMutability":"payable","type":"function","payable":true},{"inputs":[{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"withdraw","outputs":[],"stateMutability":"nonpayable","type":"function"}]`))
	if err != nil {
		log.Println("Failed to load contract ABI: %v", err)
		return "", err
	}
	fmt.Println("Contract ABI: ", contractABI)

	// 创建交易数据
	data, err := contractABI.Pack("deposit")
	if err != nil {
		log.Println("Failed to pack data: %v", err)
		return "", err
	}

	// 构建交易
	var gasLimit uint64 = 6721975
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Println("Failed to get nonce: %v", err)
		return "", err
	}
	// 将存入金额作为附加的 value 发送到合约
	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(contractAddr),
		bigTrxAmount,
		gasLimit,
		gasPrice,
		data,
	)

	// 获取 Chain ID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Println("Failed to get chain ID: %v", err)
		return "", err
	}
	fmt.Println("Connected to chain ID:", chainID)

	// 签署交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), key)
	if err != nil {
		log.Println("Failed to sign transaction: %v", err)
		return "", err
	}

	// 尝试多次发送交易
	for i := 0; i < retryAttempts; i++ {
		// 发送交易
		err = client.SendTransaction(ctx, signedTx)
		if err == nil {
			fmt.Printf("Transaction hash: %s\n", signedTx.Hash().Hex())
			return signedTx.Hash().Hex(), nil
		}

		fmt.Printf("Attempt %d failed: %v\n", i+1, err)
		time.Sleep(retryDelay)
	}

	fmt.Println("Failed to send transaction after multiple attempts.")
	err = fmt.Errorf("Failed to send transaction after multiple attempts")
	return "", err
}

func WithDrawTransaction(contractAddr string, transactionAmount string, privateKey string) (string, error) {
	ganacheURL := os.Getenv("GANACHE_URL")
	bigTrxAmount := new(big.Int)
	bigTrxAmount, success := bigTrxAmount.SetString(transactionAmount, 10)
	if !success {
		err := fmt.Errorf("Invalid token amount: %s", transactionAmount)
		return "", err
	}
	// 创建一个有效的上下文
	ctx := context.Background()

	// 连接到 Ganache 节点
	client, err := ethclient.DialContext(ctx, ganacheURL)
	if err != nil {
		log.Println("Failed to connect to the Ethereum client: %v", err)
		return "", err
	}

	fmt.Println("Successfully connected to Ganache: ", client)

	// 加载私钥
	var privatekey = strings.TrimPrefix(privateKey, "0x") // 移除前缀
	key, err := crypto.HexToECDSA(privatekey)
	if err != nil {
		log.Println("Failed to load private key: %v", err)
		return "", err
	}

	// 创建一个授权者
	fromAddress := crypto.PubkeyToAddress(key.PublicKey)
	fmt.Println("From Address: ", fromAddress.Hex())

	// 获取当前的 Gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Println("Failed to get gas price: %v", err)
		return "", err
	}
	fmt.Println("Gas price: ", gasPrice)

	// 加载合约 ABI
	contractABI, err := abi.JSON(strings.NewReader(`[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"balances","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function","constant":true},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function","constant":true},{"inputs":[],"name":"deposit","outputs":[],"stateMutability":"payable","type":"function","payable":true},{"inputs":[{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"withdraw","outputs":[],"stateMutability":"nonpayable","type":"function"}]`))
	if err != nil {
		log.Println("Failed to load contract ABI: %v", err)
		return "", err
	}
	fmt.Println("Contract ABI: ", contractABI)

	// 创建交易数据
	data, err := contractABI.Pack("withdraw", bigTrxAmount)
	if err != nil {
		log.Println("Failed to pack data: %v", err)
		return "", err
	}

	// 构建交易
	var gasLimit uint64 = 6721975
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Println("Failed to get nonce: %v", err)
		return "", err
	}
	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(contractAddr),
		nil,
		gasLimit,
		gasPrice,
		data,
	)

	// 获取 Chain ID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Println("Failed to get chain ID: %v", err)
		return "", err
	}
	fmt.Println("Connected to chain ID:", chainID)

	// 签署交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), key)
	if err != nil {
		log.Println("Failed to sign transaction: %v", err)
		return "", err
	}

	// 尝试多次发送交易
	for i := 0; i < retryAttempts; i++ {
		// 发送交易
		err = client.SendTransaction(ctx, signedTx)
		if err == nil {
			fmt.Printf("Transaction hash: %s\n", signedTx.Hash().Hex())
			return signedTx.Hash().Hex(), nil
		}

		fmt.Printf("Attempt %d failed: %v\n", i+1, err)
		time.Sleep(retryDelay)
	}

	fmt.Println("Failed to send transaction after multiple attempts.")
	err = fmt.Errorf("Failed to send transaction after multiple attempts")
	return "", err
}

func GetBalance(contractAddr string, privateKey string) (*big.Int, error) {
	ganacheURL := os.Getenv("GANACHE_URL")
	// 创建一个有效的上下文
	ctx := context.Background()
	var intValue int64 = 0
	bigValue := big.NewInt(intValue)

	// 连接到 Ganache 节点
	client, err := ethclient.DialContext(ctx, ganacheURL)
	if err != nil {
		log.Println("Failed to connect to the Ethereum client: %v", err)
		return bigValue, err
	}

	fmt.Println("Successfully connected to Ganache: ", client)

	// 加载私钥
	var privatekey = strings.TrimPrefix(privateKey, "0x") // 移除前缀
	key, err := crypto.HexToECDSA(privatekey)
	if err != nil {
		log.Println("Failed to load private key: %v", err)
		return bigValue, err
	}

	// 创建一个授权者
	fromAddress := crypto.PubkeyToAddress(key.PublicKey)
	fmt.Println("From Address: ", fromAddress.Hex())

	balance, err := client.BalanceAt(ctx, fromAddress, nil) // 第三个参数为区块号，nil 表示最新区块
	if err != nil {
		log.Println("Failed to retrieve balance: %v", err)
		return bigValue, err
	}
	log.Printf("Balance of address %s: %s wei\n", fromAddress.Hex(), balance.String())
	return balance, nil
}
