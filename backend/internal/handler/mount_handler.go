package handler

import (
	"context"
	"errors"
	"net/http"
	"openbridge/backend/internal/domain/entity"
	"openbridge/backend/internal/pkg/logger"
	"openbridge/backend/internal/pkg/myerror"
	"openbridge/backend/internal/tool"
	"openbridge/backend/internal/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MountHandler struct {
	mountUseCase *usecase.MountUseCase
}

func NewMountHandler(mountUseCase *usecase.MountUseCase) *MountHandler {
	return &MountHandler{mountUseCase: mountUseCase}
}

// CreateMount 创建挂载点
func (h *MountHandler) CreateMount(c *gin.Context) {
	var req entity.MountPoint
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, myerror.ErrorCodeJsonFormatInvalid)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}

	mount, err := h.mountUseCase.CreateMount(context.Background(), req)
	if err != nil {
		status, code := mapMountError(err)
		c.JSON(status, tool.HttpResult{Code: code, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, code)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: mount})
}

// ListMounts 列出所有挂载点
func (h *MountHandler) ListMounts(c *gin.Context) {
	mounts, err := h.mountUseCase.ListAllMounts(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, tool.HttpResult{Code: myerror.ErrorCodeMountGetFailed, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, myerror.ErrorCodeMountGetFailed)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: mounts})
}

// UpdateMount 更新挂载点
func (h *MountHandler) UpdateMount(c *gin.Context) {
	mountID, err := parseMountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, myerror.ErrorCodeParameterInvalid)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}

	var req entity.MountPoint
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, myerror.ErrorCodeJsonFormatInvalid)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}
	req.ID = mountID

	mount, err := h.mountUseCase.UpdateMount(context.Background(), req)
	if err != nil {
		status, code := mapMountError(err)
		c.JSON(status, tool.HttpResult{Code: code, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, code)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: mount})
}

// DeleteMount 删除挂载点
func (h *MountHandler) DeleteMount(c *gin.Context) {
	mountID, err := parseMountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, myerror.ErrorCodeParameterInvalid)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}

	if err := h.mountUseCase.DeleteMount(context.Background(), mountID); err != nil {
		status, code := mapMountError(err)
		c.JSON(status, tool.HttpResult{Code: code, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, code)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage})
}

// GetMountQuota 获取挂载点配额
func (h *MountHandler) GetMountQuota(c *gin.Context) {
	mountID, err := parseMountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, myerror.ErrorCodeParameterInvalid)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}

	result, err := h.mountUseCase.GetMountQuota(context.Background(), mountID)
	if err != nil {
		status, code := mapMountError(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
			code = myerror.ErrorCodeMountGetFailed
		}
		c.JSON(status, tool.HttpResult{Code: code, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, code)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: result})
}

// SyncMountQuota 同步挂载点配额
func (h *MountHandler) SyncMountQuota(c *gin.Context) {
	mountID, err := parseMountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, myerror.ErrorCodeParameterInvalid)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}

	result, err := h.mountUseCase.SyncMountQuota(context.Background(), mountID)
	if err != nil {
		status, code := mapMountError(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
			code = myerror.ErrorCodeMountGetFailed
		}
		if code == myerror.ErrorCodeQuotaResolveFailed {
			code = myerror.ErrorCodeMountQuotaSyncFailed
		}
		c.JSON(status, tool.HttpResult{Code: code, Message: err.Error()})
		c.Set(logger.LoggerErrorCodeKey, code)
		c.Set(logger.LoggerMessageKey, err.Error())
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: result})
}

func parseMountID(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	parsed, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

func mapMountError(err error) (int, int) {
	switch {
	case errors.Is(err, usecase.ErrMountInvalidMode),
		errors.Is(err, usecase.ErrMountProviderRequired),
		errors.Is(err, usecase.ErrMountParentRequired),
		errors.Is(err, usecase.ErrMountParentNotReal),
		errors.Is(err, usecase.ErrMountCircularInherit),
		errors.Is(err, usecase.ErrMountVirtualExceedsAllowed),
		errors.Is(err, usecase.ErrMountVirtualUsedInvalid),
		errors.Is(err, usecase.ErrMountDisabled),
		errors.Is(err, usecase.ErrMountDeleteInherited):
		return http.StatusBadRequest, myerror.ErrorCodeMountValidationFailed
	default:
		return http.StatusInternalServerError, myerror.ErrorCodeQuotaResolveFailed
	}
}
