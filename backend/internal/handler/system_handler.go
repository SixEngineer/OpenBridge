package handler

import (
	"net/http"

	"openbridge/backend/internal/pkg/myerror"
	"openbridge/backend/internal/tool"
	"openbridge/backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	systemUseCase *usecase.SystemUseCase
}

func NewSystemHandler(systemUseCase *usecase.SystemUseCase) *SystemHandler {
	return &SystemHandler{
		systemUseCase: systemUseCase,
	}
}

func (h *SystemHandler) PickLocalPath(c *gin.Context) {
	var req usecase.PickLocalPathInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		return
	}

	result, err := h.systemUseCase.PickLocalPath(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: result})
}

func (h *SystemHandler) GetSystemMetrics(c *gin.Context) {
	result, err := h.systemUseCase.GetSystemMetrics()
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: result})
}

func (h *SystemHandler) RestartApplication(c *gin.Context) {
	result := h.systemUseCase.RestartApplication()
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: result})
}

func (h *SystemHandler) ExitApplication(c *gin.Context) {
	result := h.systemUseCase.ExitApplication()
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: result})
}
