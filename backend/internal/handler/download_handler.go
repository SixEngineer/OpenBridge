package handler

import (
	"net/http"
	"openbridge/backend/internal/pkg/myerror"
	"openbridge/backend/internal/tool"
	"openbridge/backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type DownloadHandler struct {
	downloadUseCase *usecase.DownloadUseCase
	storageUseCase  *usecase.StorageUseCase
}

type ResolveRequest struct {
	Path string `json:"path"`
}

type CreateTaskRequest struct {
	Path string `json:"path"`
	Dir  string `json:"dir"`
}

func NewDownloadHandler(downloadUseCase *usecase.DownloadUseCase, storageUseCase *usecase.StorageUseCase) *DownloadHandler {
	return &DownloadHandler{
		downloadUseCase: downloadUseCase,
		storageUseCase:  storageUseCase,
	}
}

func (h *DownloadHandler) ResolveDirectLink(c *gin.Context) {
	var req ResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		return
	}

	result, err := h.storageUseCase.ResolveDirectLink(req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeDownloadResolveFailed, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: result})
}

func (h *DownloadHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		return
	}

	task, err := h.downloadUseCase.CreateTask(req.Path, req.Dir)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeDownloadCreateFailed, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: task})
}

func (h *DownloadHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	task, err := h.downloadUseCase.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeDownloadGetFailed, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: task})
}

func (h *DownloadHandler) GetAria2Status(c *gin.Context) {
	version, err := h.downloadUseCase.CheckAria2Status()
	if err != nil {
		c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeDownloadGetFailed, Message: err.Error(), Data: nil})
		return
	}
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: version})
}
