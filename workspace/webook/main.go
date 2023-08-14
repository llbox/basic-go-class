package main

import (
	"basic-go-class/workspace/webook/internal/repository"
	"basic-go-class/workspace/webook/internal/repository/dao"
	"basic-go-class/workspace/webook/internal/service"
	"basic-go-class/workspace/webook/internal/web"
	"basic-go-class/workspace/webook/internal/web/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func main() {
	db := initDB()
	u := initUser(db)
	server := initWebServer()
	u.RegisterUserRoutes(server)
	server.Run(":8080")
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		MaxAge:           1 * time.Hour,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "localhost") {
				return true
			}
			return strings.Contains(origin, "company.com")
		},
	}))
	//store := cookie.NewStore([]byte("secret"))
	store, err := redis.NewStore(16, "tcp", "localhost:16379", "",
		[]byte("efbNSXWCJr94OauKRHbtUVJyWynEenYW"), []byte("IZpTbmLUUHzpCks6ovsjkMiueTGYtylf"))

	if err != nil {
		panic(err)
	}
	server.Use(sessions.Sessions("mysession", store))
	server.Use(middleware.NewLoginMiddlewareBuilder().
		IgnorePaths("/users/signUp").
		IgnorePaths("/users/signIn").
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
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13306)/webook"))
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
