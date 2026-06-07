package handler

import (
	"net/http"
	"openbridge/backend/internal/pkg/myerror"
	"openbridge/backend/internal/tool"
	"openbridge/backend/internal/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
)

type StorageHandler struct {
	storageUseCase *usecase.StorageUseCase
}

func NewStorageHandler(storageUseCase *usecase.StorageUseCase) *StorageHandler {
    return &StorageHandler{
        storageUseCase: storageUseCase,
    }
}

// 获取驱动列表
func (h *StorageHandler) GetDrivers(c *gin.Context) {

	// 调用 usecase 获取驱动列表
	drivers, err := h.storageUseCase.GetDrivers()
	if err != nil {
	    c.JSON(http.StatusInternalServerError, tool.HttpResult{Code: myerror.ErrorCodeGetDriversFailed, Message: err.Error()})
		return
	}

	// 返回获取驱动列表成功的结果
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: drivers})
}

// 获取驱动信息
func (h *StorageHandler) GetDriverInfo(c *gin.Context) {
    
	// 从 URL 参数中获取驱动名称
	driverName := c.Query("name")

	// 调用 usecase 获取驱动信息
	driverInfo, err := h.storageUseCase.GetDriverInfo(driverName)
	if err != nil {
	    c.JSON(http.StatusInternalServerError, tool.HttpResult{Code: myerror.ErrorCodeGetDriverInfoFailed, Message: err.Error()})
		return
	}

	// 返回获取驱动信息成功的结果
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: driverInfo})
}

// 获取当前目录的文件列表
func (h *StorageHandler) GetFiles(c *gin.Context) {
    
	// 从 URL 参数中获取驱动名称和目录路径
	path := c.Query("path")
	page_str := c.Query("page")
	per_page_str := c.Query("per_page")

	// 将分页参数转换为 uint 类型，如果转换失败则使用默认值
	page, err := strconv.ParseUint(page_str, 10, 32)
	if err != nil {
		page = 1
	}
	per_page, err := strconv.ParseUint(per_page_str, 10, 32)
	if err != nil {
		per_page = 10
	}

	// 调用 usecase 获取文件列表
	files, err := h.storageUseCase.GetFiles(path, uint(page), uint(per_page))
	if err != nil {
	    c.JSON(http.StatusInternalServerError, tool.HttpResult{Code: myerror.ErrorCodeGetFilesFailed, Message: err.Error()})
		return
	}

	// 返回获取文件列表成功的结果
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: files})
}

// 获取文件信息
func (h *StorageHandler) GetFileInfo(c *gin.Context) {
    
	// 从 URL 参数中获取驱动名称和文件路径
	path := c.Query("path")

	// 调用 usecase 获取文件信息
	fileInfo, err := h.storageUseCase.GetFileInfo(path)
	if err != nil {
	    c.JSON(http.StatusInternalServerError, tool.HttpResult{Code: myerror.ErrorCodeGetFileInfoFailed, Message: err.Error()})
		return
	}
	// 返回获取文件信息成功的结果
	c.JSON(http.StatusOK, tool.HttpResult{Code: myerror.ErrorCodeOK, Message: myerror.SuccessMessage, Data: fileInfo})
}