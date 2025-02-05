package ganache_connector

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

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

func DepositTransaction(contractAddr string, depositAmount string, privateKey string) (string, error) {
	ganacheURL := os.Getenv("GANACHE_URL")
	bigTrxAmount := new(big.Int)
	bigTrxAmount, success := bigTrxAmount.SetString(depositAmount, 10)
	if !success {
		err := fmt.Errorf("Invalid token amount: %s", depositAmount)
		return "", err
	}
	ctx := context.Background()

	client, err := ethclient.DialContext(ctx, ganacheURL)
	if err != nil {
		log.Println("Failed to connect to the Ethereum client: %v", err)
		return "", err
	}

	fmt.Println("Successfully connected to Ganache: ", client)

	var privatekey = strings.TrimPrefix(privateKey, "0x") // remove suffix
	key, err := crypto.HexToECDSA(privatekey)
	if err != nil {
		log.Println("Failed to load private key: %v", err)
		return "", err
	}

	fromAddress := crypto.PubkeyToAddress(key.PublicKey)
	fmt.Println("From Address: ", fromAddress.Hex())

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Println("Failed to get gas price: %v", err)
		return "", err
	}
	fmt.Println("Gas price: ", gasPrice)

	contractABI, err := abi.JSON(strings.NewReader(`[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"balances","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function","constant":true},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function","constant":true},{"inputs":[],"name":"deposit","outputs":[],"stateMutability":"payable","type":"function","payable":true},{"inputs":[{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"withdraw","outputs":[],"stateMutability":"nonpayable","type":"function"}]`))
	if err != nil {
		log.Println("Failed to load contract ABI: %v", err)
		return "", err
	}
	fmt.Println("Contract ABI: ", contractABI)

	data, err := contractABI.Pack("deposit")
	if err != nil {
		log.Println("Failed to pack data: %v", err)
		return "", err
	}

	var gasLimit uint64 = 6721975
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Println("Failed to get nonce: %v", err)
		return "", err
	}
	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(contractAddr),
		bigTrxAmount,
		gasLimit,
		gasPrice,
		data,
	)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Println("Failed to get chain ID: %v", err)
		return "", err
	}
	fmt.Println("Connected to chain ID:", chainID)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), key)
	if err != nil {
		log.Println("Failed to sign transaction: %v", err)
		return "", err
	}

	err = client.SendTransaction(ctx, signedTx)
	if err == nil {
		fmt.Printf("Transaction hash: %s\n", signedTx.Hash().Hex())
		return signedTx.Hash().Hex(), nil
	}

	fmt.Println(err.Error())
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
	ctx := context.Background()

	client, err := ethclient.DialContext(ctx, ganacheURL)
	if err != nil {
		log.Println("Failed to connect to the Ethereum client: %v", err)
		return "", err
	}

	fmt.Println("Successfully connected to Ganache: ", client)

	var privatekey = strings.TrimPrefix(privateKey, "0x")
	key, err := crypto.HexToECDSA(privatekey)
	if err != nil {
		log.Println("Failed to load private key: %v", err)
		return "", err
	}

	fromAddress := crypto.PubkeyToAddress(key.PublicKey)
	fmt.Println("From Address: ", fromAddress.Hex())

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Println("Failed to get gas price: %v", err)
		return "", err
	}
	fmt.Println("Gas price: ", gasPrice)

	// load contract ABI
	contractABI, err := abi.JSON(strings.NewReader(`[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"balances","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function","constant":true},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function","constant":true},{"inputs":[],"name":"deposit","outputs":[],"stateMutability":"payable","type":"function","payable":true},{"inputs":[{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"withdraw","outputs":[],"stateMutability":"nonpayable","type":"function"}]`))
	if err != nil {
		log.Println("Failed to load contract ABI: %v", err)
		return "", err
	}
	fmt.Println("Contract ABI: ", contractABI)

	data, err := contractABI.Pack("withdraw", bigTrxAmount)
	if err != nil {
		log.Println("Failed to pack data: %v", err)
		return "", err
	}

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

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Println("Failed to get chain ID: %v", err)
		return "", err
	}
	fmt.Println("Connected to chain ID:", chainID)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), key)
	if err != nil {
		log.Println("Failed to sign transaction: %v", err)
		return "", err
	}

	err = client.SendTransaction(ctx, signedTx)
	if err == nil {
		fmt.Printf("Transaction hash: %s\n", signedTx.Hash().Hex())
		return signedTx.Hash().Hex(), nil
	}

	log.Println(err.Error())
	return "", err
}

func GetBalance(contractAddr string, privateKey string) (*big.Int, error) {
	ganacheURL := os.Getenv("GANACHE_URL")
	ctx := context.Background()
	var intValue int64 = 0
	bigValue := big.NewInt(intValue)

	client, err := ethclient.DialContext(ctx, ganacheURL)
	if err != nil {
		log.Println("Failed to connect to the Ethereum client: %v", err)
		return bigValue, err
	}

	fmt.Println("Successfully connected to Ganache: ", client)

	var privatekey = strings.TrimPrefix(privateKey, "0x") // remove suffix
	key, err := crypto.HexToECDSA(privatekey)
	if err != nil {
		log.Println("Failed to load private key: %v", err)
		return bigValue, err
	}

	fromAddress := crypto.PubkeyToAddress(key.PublicKey)
	fmt.Println("From Address: ", fromAddress.Hex())

	balance, err := client.BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		log.Println("Failed to retrieve balance: %v", err)
		return bigValue, err
	}
	log.Printf("Balance of address %s: %s wei\n", fromAddress.Hex(), balance.String())
	return balance, nil
}
