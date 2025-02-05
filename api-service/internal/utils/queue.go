package utils

import (
	domain "hw-app/internal/domain"
	"sync"
)

// task queue（channel）
var TaskQueue chan domain.Task

var TransactionQueue chan domain.Transaction

// Map that stores response channel
var ResponseMap sync.Map
