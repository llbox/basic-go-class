package middleware

import (
	"encoding/gob"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (l *LoginMiddlewareBuilder) IgnorePaths(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		gob.Register(time.Time{})
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		sess := sessions.Default(ctx)
		userId := sess.Get("userId")
		if userId == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		sess.Options(sessions.Options{
			MaxAge: 60,
		})

		now := time.Now()
		updateTime := sess.Get("update_time")
		if updateTime == nil {
			sess.Set("update_time", now)
			sess.Save()
			return
		}

		uTime := updateTime.(time.Time)
		if now.Sub(uTime) > time.Second*30 {
			sess.Set("update_time", now)
			sess.Save()
			return
		}
	}
}
