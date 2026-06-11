package handler

import (
	"io"
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

type UserBackupRequest struct {
	Password string `json:"password"`
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
	var req UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		return
	}

	deviceID := c.GetHeader(usecase.DeviceIDHeader)
	token, err := h.userUseCase.Login(req.Username, req.Password, deviceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeLoginFailed, Message: err.Error()})
		return
	}

	h.adminChecker.SetToken(token)
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

func (h *UserHandler) Backup(c *gin.Context) {
	var req UserBackupRequest
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
			return
		}
	}

	backupBytes, filename, err := h.userUseCase.Backup(req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeBackupFailed, Message: err.Error()})
		return
	}

	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/json; charset=utf-8", backupBytes)
}

func (h *UserHandler) Restore(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeRestoreFailed, Message: err.Error()})
		return
	}
	defer src.Close()

	backupBytes, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeRestoreFailed, Message: err.Error()})
		return
	}

	result, err := h.userUseCase.Restore(backupBytes, c.PostForm("password"))
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeRestoreFailed, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: result})
}

// 获取用户数据
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userInfo, err := h.userUseCase.GetUserInfo(c.GetHeader(usecase.DeviceIDHeader))
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeGetUserInfoFailed, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: userInfo})
}

func (h *UserHandler) GetSessionStatus(c *gin.Context) {
	status := h.userUseCase.GetSessionStatus(c.GetHeader(usecase.DeviceIDHeader))
	c.JSON(http.StatusOK, tool.HttpResult{
		Code:    myerror.ErrorCodeOK,
		Message: myerror.SuccessMessage,
		Data:    status,
	})
}
