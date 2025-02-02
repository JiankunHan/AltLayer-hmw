package main

import (
	"fmt"
	domain "hw-app/internal/domain"
	mysql_connector "hw-app/internal/repository"
	service "hw-app/internal/service"
	utils "hw-app/internal/utils"
	"log"
	"net/http"
	"sync"

	handler "hw-app/internal/handler"
)

// configs read from config.yaml
var config *utils.Config

// task queue（channel）
var TaskQueue chan domain.Task
var ResultQueue chan domain.Result
var TransactionQueue chan domain.Transaction

func main() {
	// load configs
	var err error
	config, err = utils.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// RequestHandler thread, deals with tasks in TaskQueue
	TaskQueue = make(chan domain.Task, config.Queue.BufferSize)
	ResultQueue = make(chan domain.Result, config.Queue.BufferSize)
	TransactionQueue = make(chan domain.Transaction, config.Queue.BufferSize)
	var wg sync.WaitGroup
	for i := 0; i < config.ReqHandler.NumReqHandlers; i++ {
		wg.Add(1)
		DB, err := mysql_connector.IntializeDBConn()
		if err != nil {
			break
		}
		handler.PoolDB = append(handler.PoolDB, DB)
		go handler.RequestHandler(i, &wg)
	}

	// Result handler thread, deals with ResultQueue and return response
	go handler.ResultHandler()

	// Chain connector thread, deals with TransactionQueue and interacts with the chain
	go handler.GanacheHandler()

	//main thread, process http requests and put them in the TaskQueue
	http.HandleFunc("/tokenClaim", service.HandleClaimRequest)

	// start HTTP service
	port := fmt.Sprintf(":%d", config.Server.Port)
	fmt.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))

	// wait till all ReqHandlers are done with tasks
	wg.Wait()
}
