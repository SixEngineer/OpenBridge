package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"openbridge/backend/internal/pkg/myerror"
	"openbridge/backend/internal/tool"
	"openbridge/backend/internal/usecase"
	"path"
	"strings"

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
	deviceID := requestDeviceID(c)
	var req ResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		return
	}

	result, err := h.storageUseCase.ResolveDirectLinkForClientDevice(deviceID, req.Path, requestOriginForDownload(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeDownloadResolveFailed, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: result})
}

func (h *DownloadHandler) DownloadDirect(c *gin.Context) {
	deviceID := requestDeviceID(c)
	path := strings.TrimSpace(c.Query("path"))
	if path == "" {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: "path is required"})
		return
	}

	result, err := h.storageUseCase.ResolveDirectLinkForDevice(deviceID, path)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeDownloadResolveFailed, Message: err.Error()})
		return
	}

	if !result.IsOpenListProxy {
		c.Redirect(http.StatusFound, result.DirectLink)
		return
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, result.DirectLink, nil)
	if err != nil {
		c.JSON(http.StatusBadGateway, tool.HttpResult{Code: myerror.ErrorCodeDownloadResolveFailed, Message: err.Error()})
		return
	}

	copyIfPresent(c.Request.Header, req.Header, "Range")
	copyIfPresent(c.Request.Header, req.Header, "If-Range")
	copyIfPresent(c.Request.Header, req.Header, "If-Modified-Since")
	copyIfPresent(c.Request.Header, req.Header, "If-None-Match")

	client := &http.Client{Timeout: 0}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, tool.HttpResult{Code: myerror.ErrorCodeDownloadResolveFailed, Message: err.Error()})
		return
	}
	defer resp.Body.Close()

	copyDownloadHeaders(resp.Header, c.Writer.Header())
	if c.Writer.Header().Get("Content-Disposition") == "" && strings.TrimSpace(result.Name) != "" {
		c.Writer.Header().Set("Content-Disposition", `attachment; filename="`+result.Name+`"`)
	}

	c.Status(resp.StatusCode)
	if c.Request.Method == http.MethodHead {
		return
	}
	_, _ = io.Copy(c.Writer, resp.Body)
}

func (h *DownloadHandler) DownloadFolderZip(c *gin.Context) {
	deviceID := requestDeviceID(c)
	folderPath := strings.TrimSpace(c.Query("path"))
	if folderPath == "" {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeParameterInvalid, Message: "path is required"})
		return
	}

	filename := path.Base(strings.TrimRight(folderPath, "/"))
	if filename == "" || filename == "." || filename == "/" {
		filename = "folder"
	}
	filename += ".zip"
	escaped := url.PathEscape(filename)
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, escaped))
	c.Header("Cache-Control", "private, max-age=60")

	if _, err := h.storageUseCase.StreamFolderZipForDevice(c.Request.Context(), deviceID, folderPath, c.Writer); err != nil {
		_ = c.Error(err)
	}
}

func (h *DownloadHandler) CreateTask(c *gin.Context) {
	deviceID := requestDeviceID(c)
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeJsonFormatInvalid, Message: err.Error()})
		return
	}

	task, err := h.downloadUseCase.CreateTask(deviceID, req.Path, req.Dir)
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

func (h *DownloadHandler) OpenFileLocation(c *gin.Context) {
	taskID := c.Param("id")
	filePath, err := h.downloadUseCase.OpenFileLocation(taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeDownloadOpenFailed, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: gin.H{"folder_path": filePath}})
}

func (h *DownloadHandler) OpenFile(c *gin.Context) {
	taskID := c.Param("id")
	filePath, err := h.downloadUseCase.OpenFile(taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeDownloadOpenFailed, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: gin.H{"file_path": filePath}})
}

func (h *DownloadHandler) RetryTask(c *gin.Context) {
	deviceID := requestDeviceID(c)
	taskID := c.Param("id")
	task, err := h.downloadUseCase.RetryTask(deviceID, taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, tool.HttpResult{Code: myerror.ErrorCodeDownloadCreateFailed, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: task})
}

func requestDeviceID(c *gin.Context) string {
	if deviceID := strings.TrimSpace(c.GetHeader(usecase.DeviceIDHeader)); deviceID != "" {
		return deviceID
	}
	return strings.TrimSpace(c.Query("device_id"))
}

func requestOriginForDownload(c *gin.Context) string {
	if forwarded := strings.TrimSpace(c.Request.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		proto := strings.TrimSpace(strings.Split(forwarded, ",")[0])
		return proto + "://" + c.Request.Host
	}
	if c.Request.TLS != nil {
		return "https://" + c.Request.Host
	}
	return "http://" + c.Request.Host
}

func copyIfPresent(src http.Header, dst http.Header, key string) {
	if value := strings.TrimSpace(src.Get(key)); value != "" {
		dst.Set(key, value)
	}
}

func copyDownloadHeaders(src http.Header, dst http.Header) {
	for _, key := range []string{
		"Accept-Ranges",
		"Content-Disposition",
		"Content-Encoding",
		"Content-Length",
		"Content-Range",
		"Content-Type",
		"ETag",
		"Last-Modified",
		"Cache-Control",
	} {
		if value := src.Get(key); value != "" {
			dst.Set(key, value)
		}
	}

	if dst.Get("Cache-Control") == "" {
		dst.Set("Cache-Control", "private, max-age=60")
	}
}
