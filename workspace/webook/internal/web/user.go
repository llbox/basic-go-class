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
	"time"
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

	ug.POST("/signup", h.Signup)
	ug.POST("/login", h.LoginJWT)
	ug.POST("/edit", h.EditJWT)
	ug.GET("/profile", h.ProfileJWT)

	//ug.POST("/v1/signIn", h.Login)
	//ug.GET("/v1/profile", h.Profile)
}

func (h *UserHandler) Signup(c *gin.Context) {
	type signUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req signUpReq
	err := c.Bind(&req)
	if err != nil {
		return
	}

	ok, err := h.emailRegExp.MatchString(req.Email)
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
		Email:    req.Email,
		Password: req.Password,
	})
	if errors.Is(err, service.ErrUserDuplicateEmail) {
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
	sess.Options(sessions.Options{
		MaxAge: 60,
	})
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
	//log.Printf("email: %s, password: %s", req.Email, req.Password)

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
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid:       user.Id,
		UserAgent: ctx.Request.UserAgent(),
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

func (h *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{
		MaxAge: -1,
	})
	sess.Save()
	ctx.String(http.StatusOK, "退出登录成功")
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
func (h *UserHandler) EditJWT(ctx *gin.Context) {
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
	//log.Printf("request: %+v", req)
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

	claims, ok := ctx.Get("user")
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		log.Printf("calims 异常: %v", claims)
		return
	}
	uc := claims.(*UserClaims)
	err = h.svc.Edit(ctx.Request.Context(), domain.User{
		Id:           uc.Uid,
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

func (h *UserHandler) ProfileJWT(ctx *gin.Context) {

	claims := ctx.MustGet("user")
	// ！！注意是指针
	userClaims, ok := claims.(*UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "内部错误")
		log.Println("非法的claims")
		return
	}
	userId := userClaims.Uid

	user, err := h.svc.Profile(ctx.Request.Context(), userId)
	if err != nil {
		ctx.String(http.StatusOK, "内部错误")
		log.Printf("用户查询失败: %v\n", err)
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
	Uid       int64
	UserAgent string
}
