package handler

import (
	"net/http"
	"openbridge/backend/internal/middleware"
	"openbridge/backend/internal/pkg/myerror"
	"openbridge/backend/internal/tool"
	"openbridge/backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserHandler struct {
	userUseCase  *usecase.UserUseCase
	adminChecker *middleware.AdminChecker
}

func NewUserHandler(userUseCase *usecase.UserUseCase, adminChecker *middleware.AdminChecker) *UserHandler {
	return &UserHandler{
		userUseCase:  userUseCase,
		adminChecker: adminChecker,
	}
}

func (h *UserHandler) UserLogin(c *gin.Context) {

	// 解析请求参数
	var req UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		return
	}

	// 调用 usecase 处理登录逻辑（返回 token）
	token, err := h.userUseCase.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeLoginFailed, Message: err.Error()})
		return
	}

	// 将 token 设置到 AdminChecker，后续中间件可用其验证角色
	h.adminChecker.SetToken(token)

	// 返回登录成功的结果
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: gin.H{"token": token}})
}

// 重置用户数据
func (h *UserHandler) Reset(c *gin.Context) {
	scope := usecase.ResetScope(c.DefaultQuery("scope", string(usecase.ResetScopeAll)))
	err := h.userUseCase.Reset(scope)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeResetFailed, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: nil})
}

// 获取用户数据
func (h *UserHandler) GetUserInfo(c *gin.Context) {

	userInfo, err := h.userUseCase.GetUserInfo()
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeGetUserInfoFailed, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: userInfo})
}

func (h *UserHandler) GetSessionStatus(c *gin.Context) {
	status := h.userUseCase.GetSessionStatus()
	c.JSON(http.StatusOK, tool.HttpResult{
		Code:    myerror.ErrorCodeOK,
		Message: myerror.SuccessMessage,
		Data:    status,
	})
}
