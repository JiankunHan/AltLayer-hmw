package utils

import (
	domain "hw-app/internal/domain"
	"sync"
)

// task queue（channel）
var TaskQueue chan domain.Task

var ResultQueue chan domain.Result

var TransactionQueue chan domain.Transaction

// 管理 response 通道的映射表
var ResponseMap sync.Map
