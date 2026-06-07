package handler

import (
	"net/http"
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
	userUseCase *usecase.UserUseCase
}

func NewUserHandler(userUseCase *usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

func (h *UserHandler) UserLogin(c *gin.Context) {
	
	// 解析请求参数
	var req UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
	    c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
	}

	// 调用 usecase 处理登录逻辑
	err := h.userUseCase.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeLoginFailed, Message: err.Error()})
		return
	}

	// 返回登录成功的结果
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: nil})
}

// 重置用户数据
func (h *UserHandler) Reset(c *gin.Context) { 

	err := h.userUseCase.Reset()
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