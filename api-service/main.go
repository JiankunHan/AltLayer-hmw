package main

import (
	mysql_connector "hw-app/internal/repository"
	claim "hw-app/internal/tools"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化数据库
	err := mysql_connector.InitDB()
	if err != nil {
		log.Fatal("Fail to create database connection: ", err)
	}
	defer mysql_connector.CloseDB()

	// 自动迁移
	// db.AutoMigrate(&mysql_connector.User{})

	// 创建 Gin 实例
	r := gin.Default()

	// RESTful API 路由
	// r.GET("/users", mysql_connector.GetUsers)
	// r.POST("/users", mysql_connector.CreateUser)
	// r.GET("/users/:id", mysql_connector.GetUser)
	// r.PUT("/users/:id", mysql_connector.UpdateUser)
	// r.DELETE("/users/:id", mysql_connector.DeleteUser)
	// r.PUT("/transactions/deposit", ganache_connector.DepositTransaction)
	// r.PUT("/transactions/withdraw", ganache_connector.WithDrawTransaction)
	r.POST("/tokenClaim", claim.CreateClaimReq)
	r.GET("/tokenClaim/claims", claim.GetTokenClaims)
	r.POST("/approval", claim.CreateClaimApproval)
	r.GET("/approval", claim.GetClaimApproval)

	// 启动服务器
	r.Run(":8080")
}
