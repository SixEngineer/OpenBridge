package handler

import (
	"net/http"
	"openbridge/backend/internal/middleware"
	"openbridge/backend/internal/pkg/myerror"
	"openbridge/backend/internal/tool"
	"openbridge/backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	settingsUseCase *usecase.SettingsUseCase
	adminChecker    *middleware.AdminChecker
}

type UpdateOpenListSettingsRequest struct {
	BaseURL string `json:"base_url" binding:"required"`
}

type UpdateAria2SettingsRequest struct {
	RPCURL    string `json:"rpc_url"`
	Path      string `json:"path"`
	AutoStart bool   `json:"auto_start"`
}

type UpdateRcloneSettingsRequest struct {
	Path string `json:"path" binding:"required"`
}

type UpdateSessionSettingsRequest struct {
	DeviceLimit int `json:"device_limit" binding:"required"`
}

type UpdateAppSettingsRequest struct {
	AutoOpenBrowser bool `json:"auto_open_browser"`
}

type UpdateFileTreeSettingsRequest struct {
	CacheSizeKB int64 `json:"cache_size_kb" binding:"required"`
	CacheDepth  int   `json:"cache_depth" binding:"required"`
}

func NewSettingsHandler(settingsUseCase *usecase.SettingsUseCase, adminChecker *middleware.AdminChecker) *SettingsHandler {
	return &SettingsHandler{
		settingsUseCase: settingsUseCase,
		adminChecker:    adminChecker,
	}
}

func (h *SettingsHandler) GetSettings(c *gin.Context) {
	c.JSON(http.StatusOK, tool.HttpResult{
		Code:    myerror.ErrorCodeOK,
		Message: myerror.SuccessMessage,
		Data:    h.settingsUseCase.GetSettings(),
	})
}

func (h *SettingsHandler) UpdateOpenList(c *gin.Context) {
	var req UpdateOpenListSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		return
	}

	settings, err := h.settingsUseCase.UpdateOpenListBaseURL(req.BaseURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeSettingsUpdateFailed, Message: err.Error()})
		return
	}

	h.adminChecker.SetBaseURL(settings.OpenListBaseURL)

	c.JSON(http.StatusOK, tool.HttpResult{
		Code:    myerror.ErrorCodeOK,
		Message: myerror.SuccessMessage,
		Data:    settings,
	})
}

func (h *SettingsHandler) UpdateAria2(c *gin.Context) {
	var req UpdateAria2SettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		return
	}

	settings, err := h.settingsUseCase.UpdateAria2(req.RPCURL, req.Path, req.AutoStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeSettingsUpdateFailed, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{
		Code:    myerror.ErrorCodeOK,
		Message: myerror.SuccessMessage,
		Data:    settings,
	})
}

func (h *SettingsHandler) UpdateApp(c *gin.Context) {
	var req UpdateAppSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		return
	}

	settings, err := h.settingsUseCase.UpdateApp(req.AutoOpenBrowser)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeSettingsUpdateFailed, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{
		Code:    myerror.ErrorCodeOK,
		Message: myerror.SuccessMessage,
		Data:    settings,
	})
}

func (h *SettingsHandler) UpdateRclone(c *gin.Context) {
	var req UpdateRcloneSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		return
	}

	settings, err := h.settingsUseCase.UpdateRclonePath(req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeSettingsUpdateFailed, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{
		Code:    myerror.ErrorCodeOK,
		Message: myerror.SuccessMessage,
		Data:    settings,
	})
}

func (h *SettingsHandler) UpdateSession(c *gin.Context) {
	var req UpdateSessionSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		return
	}

	settings, err := h.settingsUseCase.UpdateSessionDeviceLimit(req.DeviceLimit)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeSettingsUpdateFailed, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{
		Code:    myerror.ErrorCodeOK,
		Message: myerror.SuccessMessage,
		Data:    settings,
	})
}

func (h *SettingsHandler) UpdateFileTree(c *gin.Context) {
	var req UpdateFileTreeSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		return
	}

	settings, err := h.settingsUseCase.UpdateFileTree(req.CacheSizeKB, req.CacheDepth)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeSettingsUpdateFailed, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{
		Code:    myerror.ErrorCodeOK,
		Message: myerror.SuccessMessage,
		Data:    settings,
	})
}
