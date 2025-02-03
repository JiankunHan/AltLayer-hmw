package main

import (
	"fmt"
	domain "hw-app/internal/domain"
	handler "hw-app/internal/handler"
	mysql_connector "hw-app/internal/repository"
	service "hw-app/internal/service"
	utils "hw-app/internal/utils"
	"log"
	"net/http"
	"sync"
)

// configs read from config.yaml
var config *utils.Config

func main() {
	// load configs
	var err error
	config, err = utils.LoadConfig("/app/config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// RequestHandler thread, deals with tasks in TaskQueue
	utils.TaskQueue = make(chan domain.Task, config.Queue.BufferSize)
	utils.ResultQueue = make(chan domain.Result, config.Queue.BufferSize)
	utils.TransactionQueue = make(chan domain.Transaction, config.Queue.BufferSize)

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
	// wg.Add(1)
	// go handler.ResultHandler(&wg)

	// Chain connector thread, deals with TransactionQueue and interacts with the chain
	wg.Add(1)
	go handler.GanacheHandler(&wg)

	//main thread, process http requests and put them in the TaskQueue
	http.HandleFunc("/tokenClaim", service.HandleClaimRequest)

	// start HTTP service
	port := fmt.Sprintf(":%d", config.Server.Port)
	fmt.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))

	// wait till all ReqHandlers are done with tasks
	wg.Wait()
}
