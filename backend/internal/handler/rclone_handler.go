package handler

import (
	"net/http"
	"strconv"

	"openbridge/backend/internal/middleware"
	"openbridge/backend/internal/pkg/myerror"
	"openbridge/backend/internal/tool"
	"openbridge/backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type RcloneHandler struct {
	usecase      *usecase.RcloneUseCase
	adminChecker *middleware.AdminChecker
}

func NewRcloneHandler(usecase *usecase.RcloneUseCase, adminChecker *middleware.AdminChecker) *RcloneHandler {
	return &RcloneHandler{
		usecase:      usecase,
		adminChecker: adminChecker,
	}
}

func (h *RcloneHandler) ListProfiles(c *gin.Context) {
	profiles, err := h.usecase.ListProfiles()
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: profiles})
}

func (h *RcloneHandler) CreateProfile(c *gin.Context) {
	var req usecase.RcloneProfileInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		return
	}
	profile, err := h.usecase.CreateProfile(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: profile})
}

func (h *RcloneHandler) UpdateProfile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		return
	}

	var req usecase.RcloneProfileInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		return
	}

	profile, err := h.usecase.UpdateProfile(uint(id), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: profile})
}

func (h *RcloneHandler) DeleteProfile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		return
	}

	if err := h.usecase.DeleteProfile(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: nil})
}

func (h *RcloneHandler) ApplyProfile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		return
	}

	profile, err := h.usecase.ApplyProfile(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: profile})
}

func (h *RcloneHandler) MountProfile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		return
	}

	profile, err := h.usecase.StartMount(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: profile})
}

func (h *RcloneHandler) UnmountProfile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		return
	}

	profile, err := h.usecase.StopMount(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: profile})
}
