package web

import (
	"github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserHandler struct {
	emailRegExp    *regexp2.Regexp
	passwordRegExp *regexp2.Regexp
}

func NewUserHandler() *UserHandler {
	const emailPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	const passwordPattern = "^(?=.*\\d)(?=.*[a-z])(?=.*[A-Z]).{8,10}$"
	return &UserHandler{
		emailRegExp:    regexp2.MustCompile(emailPattern, regexp2.None),
		passwordRegExp: regexp2.MustCompile(passwordPattern, regexp2.None),
	}
}

func (u *UserHandler) RegisterUserRoutes(server *gin.Engine) {
	ug := server.Group("/user")

	ug.POST("/signUp", u.signUp)
	ug.POST("/signIn", u.signIn)
}

func (u *UserHandler) signUp(c *gin.Context) {
	type signUpReq struct {
		UserName string `json:"userName"`
		Password string `json:"password"`
	}

	var req signUpReq
	err := c.Bind(&req)
	if err != nil {
		return
	}

	ok, err := u.emailRegExp.MatchString(req.UserName)
	if err != nil {
		c.String(http.StatusOK, "内部错误")
		return
	}
	if !ok {
		c.String(http.StatusOK, "邮箱格式错误")
		return
	}

	ok, err = u.passwordRegExp.MatchString(req.Password)
	if err != nil {
		c.String(http.StatusOK, "内部错误")
		return
	}
	if !ok {
		c.String(http.StatusOK, "密码必须大于8位，包含大小写字母和特殊字符")
		return
	}

	c.String(http.StatusOK, "注册成功")

}

func (u *UserHandler) signIn(c *gin.Context) {

}
