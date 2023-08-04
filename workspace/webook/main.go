package main

import (
	"basic-go-class/workspace/webook/internal/web"
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()
	userHandler := web.NewUserHandler()
	userHandler.RegisterUserRoutes(server)
	server.Run(":8080")
}
