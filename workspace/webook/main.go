package main

import (
	"basic-go-class/workspace/webook/internal/config"
	"basic-go-class/workspace/webook/internal/repository"
	"basic-go-class/workspace/webook/internal/repository/dao"
	"basic-go-class/workspace/webook/internal/service"
	"basic-go-class/workspace/webook/internal/web"
	"basic-go-class/workspace/webook/internal/web/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

func main() {
	db := initDB()
	u := initUser(db)
	server := initWebServer()
	u.RegisterUserRoutes(server)

	//server := gin.Default()
	server.GET("/hello", func(context *gin.Context) {
		context.String(http.StatusOK, "hello man!")
	})
	server.Run(":8081")
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"x-jwt-token"},
		MaxAge:           60,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "localhost") {
				return true
			}
			return strings.Contains(origin, "company.com")
		},
	}))
	//store := cookie.NewStore([]byte("secret"))
	//store, err := redis.NewStore(16, "tcp", config.Config.Redis.Addr, "",
	//	[]byte("efbNSXWCJr94OauKRHbtUVJyWynEenYW"), []byte("IZpTbmLUUHzpCks6ovsjkMiueTGYtylf"))
	//
	//if err != nil {
	//	panic(err)
	//}
	//server.Use(sessions.Sessions("mysession", store))

	//server.Use(middleware.NewLoginMiddlewareBuilder().
	//	IgnorePaths("/users/signUp").
	//	IgnorePaths("/users/v1/signIn").
	//	Build())

	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login").
		IgnorePaths("/hello").
		Build())

	return server
}

func initUser(db *gorm.DB) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	return u
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
