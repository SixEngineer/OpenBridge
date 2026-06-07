package handler

import (
	"openbridge/backend/internal/domain/entity"
	"openbridge/backend/internal/pkg/logger"
	"openbridge/backend/internal/pkg/myerror"
	"openbridge/backend/internal/tool"
	"openbridge/backend/internal/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProviderHandler struct {
	ProviderUseCase *usecase.ProviderUseCase
}

// 构造函数
func NewProviderHandler(providerUseCase *usecase.ProviderUseCase) *ProviderHandler {
	return &ProviderHandler{
		ProviderUseCase: providerUseCase,
	}
}

// 注册 Provider
func (h *ProviderHandler) RegisterProvider(c *gin.Context) {
    
	// 绑定JSON数据到Provider结构体
	var providerAccount entity.ProviderAccount
	if err := c.ShouldBindJSON(&providerAccount); err != nil {
		c.JSON(400, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, myerror.ErrorCodeJsonFormatInvalid)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}

	err := h.ProviderUseCase.RegisterProvider(providerAccount)
	if err != nil {
		c.JSON(400, tool.HttpResult{Code: myerror.ErrorCodeProviderRegisterFailed, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, myerror.ErrorCodeProviderRegisterFailed)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}
	
	c.JSON(200, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: "ok"})
}

// 删除 Provider
func (h *ProviderHandler) DeleteProvider(c *gin.Context) {

    idStr := c.Query("id")
	// 这里需要将id转换为uint类型
    idUint, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, myerror.ErrorCodeParameterInvalid)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}

	err = h.ProviderUseCase.DeleteProvider(uint(idUint))
	if err != nil {
		c.JSON(400, tool.HttpResult{Code: myerror.ErrorCodeProviderDeleteFailed, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, myerror.ErrorCodeProviderDeleteFailed)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}

    c.JSON(200, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: "ok"})
}

// 更新 Provider
func (h *ProviderHandler) UpdateProvider(c *gin.Context) {
    
	// 绑定JSON数据到Provider结构体
	var providerAccount entity.ProviderAccount
	if err := c.ShouldBindJSON(&providerAccount); err != nil {
		c.JSON(400, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, myerror.ErrorCodeJsonFormatInvalid)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}

	err := h.ProviderUseCase.UpdateProvider(providerAccount)
	if err != nil {
		c.JSON(400, tool.HttpResult{Code: myerror.ErrorCodeProviderUpdateFailed, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, myerror.ErrorCodeProviderUpdateFailed)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}

	c.JSON(200, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: "ok"})

}

// 获取 Provider
func (h *ProviderHandler) GetProvider(c *gin.Context) {

    idStr := c.Query("id")
	// 这里需要将id转换为uint类型
    idUint, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
        c.JSON(400, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
        c.Set(logger.LoggerErrorCodeKey, myerror.ErrorCodeParameterInvalid)
        c.Set(logger.LoggerMessageKey, err.Error())
		return
    }

    provider, err := h.ProviderUseCase.GetProvider(uint(idUint))
	if err != nil {
		c.JSON(400, tool.HttpResult{Code: myerror.ErrorCodeProviderGetFailed, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, myerror.ErrorCodeProviderGetFailed)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}

	c.JSON(200, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: "ok", Data: provider})
}

// 获取 Provider 列表
func (h *ProviderHandler) ListProvider(c *gin.Context) {
    
	result, err := h.ProviderUseCase.ListProvider()
	if err != nil {
		c.JSON(400, tool.HttpResult{Code: myerror.ErrorCodeProviderListFailed, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, myerror.ErrorCodeProviderListFailed)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}
	c.JSON(200, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: "ok", Data: result})
}