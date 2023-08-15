package web

import (
	"basic-go-class/workspace/webook/internal/domain"
	"basic-go-class/workspace/webook/internal/service"
	"errors"
	"github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
)

type UserHandler struct {
	svc            *service.UserService
	emailRegExp    *regexp2.Regexp
	passwordRegExp *regexp2.Regexp
	birthdayRegExp *regexp2.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	const emailPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	const passwordPattern = "^(?=.*\\d)(?=.*[a-z])(?=.*[A-Z]).{8,10}$"
	const birthdayPattern = "^((?:19[2-9]\\d{1})|(?:20(?:(?:0[0-9])|(?:1[0-8]))))\\-((?:0?[1-9])|(?:1[0-2]))\\-((?:0?[1-9])|(?:[1-2][0-9])|30|31)$"
	return &UserHandler{
		emailRegExp:    regexp2.MustCompile(emailPattern, regexp2.None),
		passwordRegExp: regexp2.MustCompile(passwordPattern, regexp2.None),
		birthdayRegExp: regexp2.MustCompile(birthdayPattern, regexp2.None),
		svc:            svc,
	}
}

func (h *UserHandler) RegisterUserRoutes(server *gin.Engine) {
	ug := server.Group("/users")

	ug.POST("/signUp", h.SignUp)
	//ug.POST("/signIn", h.Login)
	ug.POST("/signIn", h.LoginJWT)
	ug.POST("/edit", h.Edit)
	ug.GET("/profile", h.Profile)
}

func (h *UserHandler) SignUp(c *gin.Context) {
	type signUpReq struct {
		UserName        string `json:"userName"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req signUpReq
	err := c.Bind(&req)
	if err != nil {
		return
	}

	ok, err := h.emailRegExp.MatchString(req.UserName)
	if err != nil {
		c.String(http.StatusOK, "内部错误")
		return
	}
	if !ok {
		c.String(http.StatusOK, "邮箱格式错误")
		return
	}

	ok, err = h.passwordRegExp.MatchString(req.Password)
	if err != nil {
		c.String(http.StatusOK, "内部错误")
		return
	}
	if !ok {
		c.String(http.StatusOK, "密码必须大于8位，包含大小写字母和特殊字符")
		return
	}

	if req.ConfirmPassword != req.Password {
		c.String(http.StatusOK, "两次输入的密码不一致")
		return
	}

	err = h.svc.SignUp(c.Request.Context(), domain.User{
		Email:    req.UserName,
		Password: req.Password,
	})
	if err == service.ErrUserDuplicateEmail {
		c.String(http.StatusOK, "邮箱冲突")
		return
	}

	if err != nil {
		c.String(http.StatusOK, "系统异常")
		return
	}

	c.String(http.StatusOK, "注册成功")

}

func (h *UserHandler) Login(ctx *gin.Context) {
	type signInReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req signInReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	log.Printf("email: %s, password: %s", req.Email, req.Password)

	user, err := h.svc.SignIn(ctx.Request.Context(), req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "无效的账户或密码")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	sess := sessions.Default(ctx)
	sess.Set("userId", user.Id)
	sess.Save()
	ctx.String(http.StatusOK, "登录成功")
	return
}

func (h *UserHandler) LoginJWT(ctx *gin.Context) {
	type signInReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req signInReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	log.Printf("email: %s, password: %s", req.Email, req.Password)

	user, err := h.svc.SignIn(ctx.Request.Context(), req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.String(http.StatusOK, "无效的账户或密码")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 登陆成功处理token
	claims := UserClaims{
		Uid: user.Id,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("gtIa6KYnzwFIqf1Dt6z62mmdFhRNmsfw"))
	ctx.Header("x-jwt-token", tokenStr)

	if err != nil {
		return
	}
	ctx.String(http.StatusOK, "登录成功")
	return
}

func (h *UserHandler) Edit(ctx *gin.Context) {
	type editReq struct {
		Nickname     string `json:"nickname"`
		Birthday     string `json:"birthday"`
		Introduction string `json:"introduction"`
	}
	var req editReq
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "参数格式异常")
		return
	}
	log.Printf("request: %+v", req)
	// nickname 长度[5, 10]
	if len(req.Nickname) < 4 || len(req.Nickname) > 10 {
		ctx.String(http.StatusOK, "昵称长度最小为4，最大为10")
		return
	}
	// birthday yyyy-mm-dd格式校验
	ok, err := h.birthdayRegExp.MatchString(req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "生日校验系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "生日格式错误")
		return
	}
	// introduction 长度[0, 255]
	if len(req.Introduction) > 255 {
		ctx.String(http.StatusOK, "简介长度不能超过255个字符")
		return
	}

	sess := sessions.Default(ctx)
	userId := sess.Get("userId").(int64)

	err = h.svc.Edit(ctx.Request.Context(), domain.User{
		Id:           userId,
		Nickname:     req.Nickname,
		Birthday:     req.Birthday,
		Introduction: req.Introduction,
	})
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "更新成功")
	return
}

func (h *UserHandler) Profile(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	userId := sess.Get("userId").(int64)
	user, err := h.svc.Profile(ctx.Request.Context(), userId)
	if err != nil {
		ctx.String(http.StatusOK, "内部错误")
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"userId":       user.Id,
		"email":        user.Email,
		"nickName":     user.Nickname,
		"birthday":     user.Birthday,
		"introduction": user.Introduction,
	})
	return
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid int64
}
