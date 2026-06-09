package usecase

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"openbridge/backend/internal/config"
	"openbridge/backend/internal/pkg/logger"
	"path"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type StorageUseCase struct {
	config              *config.Config
	accessProvider      openListAccessProvider
	fileTreeCache       *FileTreeCache
	fileTreeWarmMu      sync.Mutex
	fileTreeWarmRunning bool
	fileTreePathMu      sync.Mutex
	fileTreePathWarm    map[string]struct{}
}

type openListAccessProvider interface {
	GetOpenListAccess(deviceID string) (OpenListAccess, error)
}

// 1. 以下一个结构体用于 Get driver names (Admin) 接口的响应解析

// 1.0 DriverResponse 定义用于解析驱动列表响应的结构体
type DriverResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    []string `json:"data"`
}

// 2. 以下四个结构体用于 Get driver info (Admin) 接口的响应解析

// 2.0
type InfoResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    InfoResponseData `json:"data"`
}

// 2.1 data 字段的结构
type InfoResponseData struct {
	Common     []ConfigField `json:"common"`
	Additional []ConfigField `json:"additional"`
	Config     StorageConfig `json:"config"`
}

// 2.2 配置字段（common 和 additional 共用）
type ConfigField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Default  string `json:"default"`
	Options  string `json:"options"`
	Required bool   `json:"required"`
	Help     string `json:"help"`
}

// 2.3 config 字段的结构
type StorageConfig struct {
	Name        string `json:"name"`
	LocalSort   bool   `json:"local_sort"`
	OnlyProxy   bool   `json:"only_proxy"`
	NoCache     bool   `json:"no_cache"`
	NoUpload    bool   `json:"no_upload"`
	NeedMs      bool   `json:"need_ms"`
	DefaultRoot string `json:"default_root"`
	Alert       string `json:"alert"`
	OnlyIndices bool   `json:"only_indices"`
	PreferProxy bool   `json:"prefer_proxy"`
}

// 3. 以下三个结构体用于 Get driver info (Admin) 接口的响应解析

// 3.0 最外层响应结构
type FileListResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    FileListData `json:"data"`
}

// 3.1 data 字段的结构
type FileListData struct {
	Content  []FileItem `json:"content"`
	Total    int        `json:"total"`
	Readme   string     `json:"readme"`
	Header   string     `json:"header"`
	Write    bool       `json:"write"`
	Provider string     `json:"provider"`
}

// 3.2 文件/目录项结构
type FileItem struct {
	Name     string      `json:"name"`
	Size     int64       `json:"size"`
	IsDir    bool        `json:"is_dir"`
	Modified time.Time   `json:"modified"`
	Created  time.Time   `json:"created"`
	Sign     string      `json:"sign"`
	Thumb    string      `json:"thumb"`
	Type     int         `json:"type"`      // 1: 目录, 可能是其他值表示文件
	Hashinfo string      `json:"hashinfo"`  // 注意原始 JSON 中是字符串 "null"
	HashInfo interface{} `json:"hash_info"` // 原始 JSON 中是 null
}

// 4. 以下一个结构体用于 Get file info (Admin) 接口的响应解析

// 4.0 最外层响应结构
type FileInfoResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    FileDetail `json:"data"`
}

// 4.1 data 字段的结构
type FileDetail struct {
	Name     string      `json:"name"`
	Size     int64       `json:"size"`
	IsDir    bool        `json:"is_dir"`
	Modified string      `json:"modified"` // 或 time.Time
	Created  string      `json:"created"`  // 或 time.Time
	Sign     string      `json:"sign"`
	Thumb    string      `json:"thumb"`
	Type     int         `json:"type"`
	Hashinfo string      `json:"hashinfo"`
	HashInfo interface{} `json:"hash_info"`
	RawURL   string      `json:"raw_url"`
	Readme   string      `json:"readme"`
	Header   string      `json:"header"`
	Provider string      `json:"provider"`
	Related  interface{} `json:"related"`
}

type DirectLinkResult struct {
	Path            string `json:"path"`
	Name            string `json:"name"`
	Size            int64  `json:"size"`
	Provider        string `json:"provider"`
	DirectLink      string `json:"direct_link"`
	Header          string `json:"header,omitempty"`
	IsOpenListProxy bool   `json:"is_openlist_proxy"`
	StoragePath     string `json:"-"`
}

type FileOperationResult struct {
	Operation string   `json:"operation"`
	Dir       string   `json:"dir,omitempty"`
	Path      string   `json:"path,omitempty"`
	DstDir    string   `json:"dst_dir,omitempty"`
	Names     []string `json:"names,omitempty"`
}

type zipFileSource struct {
	Name       string
	DirectLink string
	Header     string
}

func NewStorageUseCase(config *config.Config, accessProvider openListAccessProvider) *StorageUseCase {
	usecase := &StorageUseCase{
		config:           config,
		accessProvider:   accessProvider,
		fileTreeCache:    NewFileTreeCache(config),
		fileTreePathWarm: make(map[string]struct{}),
	}
	if err := usecase.fileTreeCache.Load(); err != nil {
		logger.L().Warn("load filetree cache failed", zap.Error(err))
	}
	usecase.warmFileTreeCacheAsync()
	return usecase
}

func (s *StorageUseCase) defaultOpenListAccess() OpenListAccess {
	return OpenListAccess{
		BaseURL: strings.TrimRight(strings.TrimSpace(s.config.OpenList.BaseURL), "/"),
		Token:   strings.TrimSpace(s.config.OpenList.Token),
	}
}

func (s *StorageUseCase) getOpenListAccess(deviceID string) (OpenListAccess, error) {
	if strings.TrimSpace(deviceID) == "" || s.accessProvider == nil {
		return s.defaultOpenListAccess(), nil
	}

	access, err := s.accessProvider.GetOpenListAccess(deviceID)
	if err != nil {
		return OpenListAccess{}, err
	}

	access.BaseURL = strings.TrimRight(strings.TrimSpace(access.BaseURL), "/")
	access.Token = strings.TrimSpace(access.Token)
	if access.BaseURL == "" {
		access.BaseURL = s.defaultOpenListAccess().BaseURL
	}
	return access, nil
}

func resolveVisiblePath(basePath string, visiblePath string) string {
	normalizedBase := normalizeStoragePath(basePath)
	normalizedVisible := normalizeStoragePath(visiblePath)

	if normalizedBase != "/" && (normalizedVisible == normalizedBase || strings.HasPrefix(normalizedVisible, normalizedBase+"/")) {
		return normalizedVisible
	}
	return normalizedVisible
}

func resolveFallbackStoragePath(basePath string, visiblePath string) string {
	normalizedBase := normalizeStoragePath(basePath)
	normalizedVisible := normalizeStoragePath(visiblePath)

	if normalizedBase == "/" {
		return normalizedVisible
	}
	if normalizedVisible == "/" {
		return normalizedBase
	}
	if normalizedVisible == normalizedBase || strings.HasPrefix(normalizedVisible, normalizedBase+"/") {
		return normalizedVisible
	}
	return normalizeStoragePath(normalizedBase + "/" + strings.TrimLeft(normalizedVisible, "/"))
}

func storagePathCandidates(basePath string, visiblePath string) []string {
	primary := resolveVisiblePath(basePath, visiblePath)
	fallback := resolveFallbackStoragePath(basePath, visiblePath)
	if fallback == primary {
		return []string{primary}
	}
	return []string{primary, fallback}
}

func shouldTryFallbackForPath(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(message, "object not found") ||
		strings.Contains(message, "failed get dir") ||
		strings.Contains(message, "failed to get obj")
}

// GetDrivers 获取所有驱动名称
func (s *StorageUseCase) GetDrivers() ([]string, error) {

	client := http.Client{Timeout: 10 * time.Second}

	// 构造HTTP GET请求，目标URL为OpenList的驱动列表接口
	req, err := http.NewRequest("GET", s.config.OpenList.BaseURL+"/api/admin/driver/names", nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头，指定User-Agent，并添加Authorization头以使用Bearer Token进行认证
	req.Header.Set("User-Agent", "OpenBridge/1.0")
	req.Header.Set("Authorization", s.config.OpenList.Token)

	// 发送HTTP请求并获取响应
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析JSON响应
	var driverResponse DriverResponse
	if err := json.Unmarshal(body, &driverResponse); err != nil {
		return nil, err
	}

	return driverResponse.Data, nil
}

// GetDriverInfo 获取指定驱动的详细信息
func (s *StorageUseCase) GetDriverInfo(driverName string) (*InfoResponseData, error) {

	client := http.Client{Timeout: 10 * time.Second}

	// 构造HTTP GET请求，目标URL为OpenList的驱动列表接口
	req, err := http.NewRequest("GET", s.config.OpenList.BaseURL+"/api/admin/driver/info?driver="+driverName, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头，指定User-Agent，并添加Authorization头以使用Bearer Token进行认证
	req.Header.Set("User-Agent", "OpenBridge/1.0")
	req.Header.Set("Authorization", s.config.OpenList.Token)

	// 发送HTTP请求并获取响应
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析JSON响应
	var InfoResponse InfoResponse
	if err := json.Unmarshal(body, &InfoResponse); err != nil {
		return nil, err
	}

	return &InfoResponse.Data, nil
}

// GetFiles 获取指定目录下的文件列表
func (s *StorageUseCase) GetFiles(path string, page uint, pageSize uint) (*FileListData, error) {
	normalizedPath := normalizeStoragePath(path)

	if s.fileTreeCache != nil {
		if cached, ok := s.fileTreeCache.Get(normalizedPath, page, pageSize); ok {
			return cached, nil
		}
		if s.fileTreeCache.ShouldCache(normalizedPath) {
			s.schedulePathCacheRefresh(normalizedPath)
		}
	}

	return s.fetchFilesRemote(normalizedPath, page, pageSize)
}

func (s *StorageUseCase) GetFilesForDevice(deviceID string, visiblePath string, page uint, pageSize uint) (*FileListData, error) {
	access, err := s.getOpenListAccess(deviceID)
	if err != nil {
		return nil, err
	}

	candidates := storagePathCandidates(access.BasePath, visiblePath)
	var lastErr error
	for index, storagePath := range candidates {
		data, err := s.fetchFilesRemoteWithAccess(access, storagePath, page, pageSize)
		if err == nil {
			return data, nil
		}
		lastErr = err
		if index < len(candidates)-1 && shouldTryFallbackForPath(err) {
			continue
		}
		return nil, err
	}
	return nil, lastErr
}

func (s *StorageUseCase) UpdateFileTreeConfig(maxBytes int64, maxDepth int) error {
	if s.fileTreeCache == nil {
		return nil
	}
	if err := s.fileTreeCache.UpdateConfig(maxBytes, maxDepth); err != nil {
		return err
	}
	s.warmFileTreeCacheAsync()
	return nil
}

func (s *StorageUseCase) UpdateOpenListSource() error {
	if s.fileTreeCache == nil {
		return nil
	}
	if err := s.fileTreeCache.SwitchSource(s.config.OpenList.BaseURL); err != nil {
		return err
	}
	s.warmFileTreeCacheAsync()
	return nil
}

func (s *StorageUseCase) FlushFileTreeCache() error {
	if s.fileTreeCache == nil {
		return nil
	}
	return s.fileTreeCache.Save()
}

func (s *StorageUseCase) fetchFilesRemote(path string, page uint, pageSize uint) (*FileListData, error) {
	return s.fetchFilesRemoteWithAccess(s.defaultOpenListAccess(), path, page, pageSize)
}

func (s *StorageUseCase) fetchAllFilesForCache(storagePath string) (*FileListData, error) {
	return s.fetchAllFilesForCacheWithAccess(s.defaultOpenListAccess(), storagePath)
}

func (s *StorageUseCase) fetchFilesRemoteWithAccess(access OpenListAccess, path string, page uint, pageSize uint) (*FileListData, error) {
	if strings.TrimSpace(access.Token) == "" {
		return nil, fmt.Errorf("openlist token is empty")
	}
	if strings.TrimSpace(access.BaseURL) == "" {
		return nil, fmt.Errorf("openlist base url is empty")
	}

	client := http.Client{Timeout: 0}
	requestBody := map[string]interface{}{
		"path":     path,
		"page":     page,
		"per_page": pageSize,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", access.BaseURL+"/api/fs/list", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "OpenBridge/1.0")
	req.Header.Set("Authorization", access.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openlist /api/fs/list returned status %d", resp.StatusCode)
	}

	var fileListResponse FileListResponse
	if err := json.NewDecoder(resp.Body).Decode(&fileListResponse); err != nil {
		return nil, err
	}

	if fileListResponse.Code != http.StatusOK {
		message := strings.TrimSpace(fileListResponse.Message)
		if message == "" {
			message = "unknown error"
		}
		return nil, fmt.Errorf("openlist /api/fs/list failed: %s", message)
	}

	return &fileListResponse.Data, nil
}

func (s *StorageUseCase) fetchAllFilesForCacheWithAccess(access OpenListAccess, storagePath string) (*FileListData, error) {
	const perPage = uint(200)

	var (
		page     uint = 1
		fetched  int
		combined = &FileListData{}
	)

	for {
		current, err := s.fetchFilesRemoteWithAccess(access, storagePath, page, perPage)
		if err != nil {
			return nil, err
		}
		if page == 1 {
			combined.Readme = current.Readme
			combined.Header = current.Header
			combined.Write = current.Write
			combined.Provider = current.Provider
			combined.Total = current.Total
		}

		combined.Content = append(combined.Content, current.Content...)
		fetched += len(current.Content)

		if len(current.Content) == 0 || fetched >= current.Total || len(current.Content) < int(perPage) {
			break
		}
		page++
	}

	if combined.Total <= 0 {
		combined.Total = len(combined.Content)
	}
	return combined, nil
}

func (s *StorageUseCase) warmFileTreeCacheAsync() {
	if s.fileTreeCache == nil || !s.fileTreeCache.Enabled() {
		return
	}

	s.fileTreeWarmMu.Lock()
	if s.fileTreeWarmRunning {
		s.fileTreeWarmMu.Unlock()
		return
	}
	s.fileTreeWarmRunning = true
	s.fileTreeWarmMu.Unlock()

	go func() {
		defer func() {
			s.fileTreeWarmMu.Lock()
			s.fileTreeWarmRunning = false
			s.fileTreeWarmMu.Unlock()
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if err := s.refreshFileTreeCachePath(ctx, "/", 0); err != nil {
			logger.L().Warn("warm filetree cache failed", zap.Error(err))
		}
		if err := s.FlushFileTreeCache(); err != nil {
			logger.L().Warn("persist warmed filetree cache failed", zap.Error(err))
		}
	}()
}

func (s *StorageUseCase) refreshFileTreeCachePath(ctx context.Context, storagePath string, depth int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if s.fileTreeCache == nil || depth > s.fileTreeCache.MaxDepth() {
		return nil
	}

	data, err := s.fetchAllFilesForCache(storagePath)
	if err != nil {
		return err
	}
	if err := s.fileTreeCache.Put(storagePath, data); err != nil {
		return err
	}
	if depth == s.fileTreeCache.MaxDepth() {
		return nil
	}

	const maxConcurrentRefresh = 2
	sem := make(chan struct{}, maxConcurrentRefresh)
	errCh := make(chan error, len(data.Content))
	var wg sync.WaitGroup

	for _, item := range data.Content {
		if !item.IsDir && item.Type != 1 {
			continue
		}

		childPath := joinStoragePath(storagePath, item.Name)
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if err := s.refreshFileTreeCachePath(ctx, path, depth+1); err != nil && ctx.Err() == nil {
				errCh <- err
			}
		}(childPath)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *StorageUseCase) schedulePathCacheRefresh(storagePath string) {
	if s.fileTreeCache == nil || !s.fileTreeCache.ShouldCache(storagePath) {
		return
	}

	s.fileTreePathMu.Lock()
	if _, exists := s.fileTreePathWarm[storagePath]; exists {
		s.fileTreePathMu.Unlock()
		return
	}
	s.fileTreePathWarm[storagePath] = struct{}{}
	s.fileTreePathMu.Unlock()

	go func(path string) {
		defer func() {
			s.fileTreePathMu.Lock()
			delete(s.fileTreePathWarm, path)
			s.fileTreePathMu.Unlock()
		}()

		data, err := s.fetchAllFilesForCache(path)
		if err != nil {
			logger.L().Warn("async filetree cache refresh failed", zap.Error(err), zap.String("path", path))
			return
		}
		if err := s.fileTreeCache.Put(path, data); err != nil {
			logger.L().Warn("async filetree cache store failed", zap.Error(err), zap.String("path", path))
			return
		}
		if err := s.FlushFileTreeCache(); err != nil {
			logger.L().Warn("persist async filetree cache failed", zap.Error(err), zap.String("path", path))
		}
	}(storagePath)
}

// GetFileInfo 获取指定文件的信息
func (s *StorageUseCase) GetFileInfo(path string) (*FileDetail, error) {
	return s.fetchFileInfoRemoteWithAccess(s.defaultOpenListAccess(), normalizeStoragePath(path))
}

func (s *StorageUseCase) GetFileInfoForDevice(deviceID string, visiblePath string) (*FileDetail, error) {
	access, err := s.getOpenListAccess(deviceID)
	if err != nil {
		return nil, err
	}
	candidates := storagePathCandidates(access.BasePath, visiblePath)
	var lastErr error
	for index, storagePath := range candidates {
		detail, err := s.fetchFileInfoRemoteWithAccess(access, storagePath)
		if err == nil {
			return detail, nil
		}
		lastErr = err
		if index < len(candidates)-1 && shouldTryFallbackForPath(err) {
			continue
		}
		return nil, err
	}
	return nil, lastErr
}

func (s *StorageUseCase) RemoveFilesForDevice(deviceID string, visibleDir string, names []string) (FileOperationResult, error) {
	cleanNames := normalizeOperationNames(names)
	if len(cleanNames) == 0 {
		return FileOperationResult{}, errors.New("no files selected")
	}

	access, err := s.getOpenListAccess(deviceID)
	if err != nil {
		return FileOperationResult{}, err
	}

	var lastErr error
	for _, storageDir := range storagePathCandidates(access.BasePath, visibleDir) {
		err := s.postOpenListFSAction(access, "remove", map[string]interface{}{
			"dir":   storageDir,
			"names": cleanNames,
		})
		if err == nil {
			s.invalidateFileTreeCache(storageDir)
			return FileOperationResult{Operation: "remove", Dir: visibleDir, Names: cleanNames}, nil
		}
		lastErr = err
		if shouldTryFallbackForPath(err) {
			continue
		}
		return FileOperationResult{}, err
	}
	return FileOperationResult{}, lastErr
}

func (s *StorageUseCase) RenameFileForDevice(deviceID string, visiblePath string, newName string) (FileOperationResult, error) {
	trimmedName := strings.TrimSpace(newName)
	if trimmedName == "" || strings.Contains(trimmedName, "/") || strings.Contains(trimmedName, "\\") {
		return FileOperationResult{}, errors.New("new name invalid")
	}

	access, err := s.getOpenListAccess(deviceID)
	if err != nil {
		return FileOperationResult{}, err
	}

	var lastErr error
	for _, storagePath := range storagePathCandidates(access.BasePath, visiblePath) {
		err := s.postOpenListFSAction(access, "rename", map[string]interface{}{
			"path": storagePath,
			"name": trimmedName,
		})
		if err == nil {
			s.invalidateFileTreeCache(path.Dir(storagePath))
			return FileOperationResult{Operation: "rename", Path: visiblePath, Names: []string{trimmedName}}, nil
		}
		lastErr = err
		if shouldTryFallbackForPath(err) {
			continue
		}
		return FileOperationResult{}, err
	}
	return FileOperationResult{}, lastErr
}

func (s *StorageUseCase) CopyFilesForDevice(deviceID string, visibleSrcDir string, visibleDstDir string, names []string) (FileOperationResult, error) {
	return s.transferFilesForDevice(deviceID, "copy", visibleSrcDir, visibleDstDir, names)
}

func (s *StorageUseCase) MoveFilesForDevice(deviceID string, visibleSrcDir string, visibleDstDir string, names []string) (FileOperationResult, error) {
	return s.transferFilesForDevice(deviceID, "move", visibleSrcDir, visibleDstDir, names)
}

func (s *StorageUseCase) transferFilesForDevice(deviceID string, operation string, visibleSrcDir string, visibleDstDir string, names []string) (FileOperationResult, error) {
	cleanNames := normalizeOperationNames(names)
	if len(cleanNames) == 0 {
		return FileOperationResult{}, errors.New("no files selected")
	}

	access, err := s.getOpenListAccess(deviceID)
	if err != nil {
		return FileOperationResult{}, err
	}

	srcCandidates := storagePathCandidates(access.BasePath, visibleSrcDir)
	dstCandidates := storagePathCandidates(access.BasePath, visibleDstDir)
	var lastErr error
	for _, srcDir := range srcCandidates {
		for _, dstDir := range dstCandidates {
			err := s.postOpenListFSAction(access, operation, map[string]interface{}{
				"src_dir": srcDir,
				"dst_dir": dstDir,
				"names":   cleanNames,
			})
			if err == nil {
				s.invalidateFileTreeCache(srcDir, dstDir)
				return FileOperationResult{Operation: operation, Dir: visibleSrcDir, DstDir: visibleDstDir, Names: cleanNames}, nil
			}
			lastErr = err
			if shouldTryFallbackForPath(err) {
				continue
			}
			return FileOperationResult{}, err
		}
	}
	return FileOperationResult{}, lastErr
}

func (s *StorageUseCase) postOpenListFSAction(access OpenListAccess, action string, payload map[string]interface{}) error {
	if strings.TrimSpace(access.Token) == "" {
		return fmt.Errorf("openlist token is empty")
	}
	if strings.TrimSpace(access.BaseURL) == "" {
		return fmt.Errorf("openlist base url is empty")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := http.Client{Timeout: 0}
	req, err := http.NewRequest("POST", access.BaseURL+"/api/fs/"+strings.Trim(action, "/"), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "OpenBridge/1.0")
	req.Header.Set("Authorization", access.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("openlist /api/fs/%s returned status %d", action, resp.StatusCode)
	}

	var openListResp struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&openListResp); err != nil {
		return err
	}
	if openListResp.Code != http.StatusOK {
		message := strings.TrimSpace(openListResp.Message)
		if message == "" {
			message = "unknown error"
		}
		return fmt.Errorf("openlist /api/fs/%s failed: %s", action, message)
	}
	return nil
}

func (s *StorageUseCase) invalidateFileTreeCache(paths ...string) {
	if s.fileTreeCache == nil {
		return
	}
	for _, p := range paths {
		if strings.TrimSpace(p) != "" {
			s.fileTreeCache.Invalidate(p)
		}
	}
}

func (s *StorageUseCase) ResolveDirectLink(path string) (*DirectLinkResult, error) {
	return s.resolveDirectLinkWithAccess(s.defaultOpenListAccess(), normalizeStoragePath(path), normalizeStoragePath(path))
}

func (s *StorageUseCase) ResolveDirectLinkForDevice(deviceID string, visiblePath string) (*DirectLinkResult, error) {
	access, err := s.getOpenListAccess(deviceID)
	if err != nil {
		return nil, err
	}
	candidates := storagePathCandidates(access.BasePath, visiblePath)
	var lastErr error
	for index, storagePath := range candidates {
		result, err := s.resolveDirectLinkWithAccess(access, normalizeStoragePath(visiblePath), storagePath)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if index < len(candidates)-1 && shouldTryFallbackForPath(err) {
			continue
		}
		return nil, err
	}
	return nil, lastErr
}

func (s *StorageUseCase) resolveDirectLinkWithAccess(access OpenListAccess, visiblePath string, storagePath string) (*DirectLinkResult, error) {
	detail, err := s.fetchFileInfoRemoteWithAccess(access, storagePath)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(detail.RawURL) == "" {
		return nil, errors.New("raw_url empty")
	}

	rawURL, err := url.Parse(detail.RawURL)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(access.BaseURL)
	if err != nil {
		return nil, err
	}
	if baseURL.Host == "" && baseURL.Path != "" {
		baseURL, err = url.Parse("http://" + access.BaseURL)
		if err != nil {
			return nil, err
		}
	}

	isProxy := false
	if rawURL.Host != "" && baseURL.Host != "" {
		isProxy = strings.EqualFold(rawURL.Host, baseURL.Host)
	}

	return &DirectLinkResult{
		Path:            visiblePath,
		Name:            detail.Name,
		Size:            detail.Size,
		Provider:        detail.Provider,
		DirectLink:      detail.RawURL,
		Header:          detail.Header,
		IsOpenListProxy: isProxy,
		StoragePath:     storagePath,
	}, nil
}

func (s *StorageUseCase) ResolveDirectLinkForClient(path string, publicBaseURL string) (*DirectLinkResult, error) {
	result, err := s.ResolveDirectLink(path)
	if err != nil {
		return nil, err
	}

	if result.IsOpenListProxy {
		base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
		if base != "" {
			result.DirectLink = base + pathJoinURL("/api/v1/download/direct") + "?path=" + url.QueryEscape(path)
		}
	}

	return result, nil
}

func (s *StorageUseCase) ResolveDirectLinkForClientDevice(deviceID string, path string, publicBaseURL string) (*DirectLinkResult, error) {
	result, err := s.ResolveDirectLinkForDevice(deviceID, path)
	if err != nil {
		return nil, err
	}

	if result.IsOpenListProxy {
		base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
		if base != "" {
			query := "?path=" + url.QueryEscape(path)
			if trimmedDeviceID := strings.TrimSpace(deviceID); trimmedDeviceID != "" {
				query += "&device_id=" + url.QueryEscape(trimmedDeviceID)
			}
			result.DirectLink = base + pathJoinURL("/api/v1/download/direct") + query
		}
	}

	return result, nil
}

func pathJoinURL(parts ...string) string {
	joined := path.Join(parts...)
	if !strings.HasPrefix(joined, "/") {
		joined = "/" + joined
	}
	return joined
}

func (s *StorageUseCase) StreamFolderZip(ctx context.Context, folderPath string, writer io.Writer) (string, error) {
	return s.streamFolderZipWithAccess(ctx, s.defaultOpenListAccess(), normalizeStoragePath(folderPath), writer)
}

func (s *StorageUseCase) StreamFolderZipForDevice(ctx context.Context, deviceID string, folderPath string, writer io.Writer) (string, error) {
	access, err := s.getOpenListAccess(deviceID)
	if err != nil {
		return "", err
	}
	candidates := storagePathCandidates(access.BasePath, folderPath)
	var lastErr error
	for index, storagePath := range candidates {
		filename, err := s.streamFolderZipWithAccess(ctx, access, storagePath, writer)
		if err == nil {
			return filename, nil
		}
		lastErr = err
		if index < len(candidates)-1 && shouldTryFallbackForPath(err) {
			continue
		}
		return "", err
	}
	return "", lastErr
}

func (s *StorageUseCase) streamFolderZipWithAccess(ctx context.Context, access OpenListAccess, folderPath string, writer io.Writer) (string, error) {
	detail, err := s.fetchFileInfoRemoteWithAccess(access, folderPath)
	if err != nil {
		return "", err
	}
	if !detail.IsDir {
		return "", fmt.Errorf("path is not a folder: %s", folderPath)
	}

	rootName := sanitizeZipName(detail.Name)
	if rootName == "" {
		rootName = sanitizeZipName(path.Base(strings.TrimRight(folderPath, "/")))
	}
	if rootName == "" || rootName == "." || rootName == "/" {
		rootName = "folder"
	}

	sources, err := s.collectFolderZipSourcesWithAccess(access, folderPath, "")
	if err != nil {
		return "", err
	}
	if len(sources) == 0 {
		return "", errors.New("folder empty")
	}

	zipWriter := zip.NewWriter(writer)
	for _, source := range sources {
		header := &zip.FileHeader{
			Name:   path.Join(rootName, source.Name),
			Method: zip.Deflate,
		}
		header.SetMode(0644)

		entryWriter, err := zipWriter.CreateHeader(header)
		if err != nil {
			return "", err
		}
		if err := s.streamRemoteFile(ctx, source.DirectLink, source.Header, entryWriter); err != nil {
			return "", err
		}
	}

	if err := zipWriter.Close(); err != nil {
		return "", err
	}

	return rootName + ".zip", nil
}

func (s *StorageUseCase) collectFolderZipSources(folderPath string, relativeDir string) ([]zipFileSource, error) {
	return s.collectFolderZipSourcesWithAccess(s.defaultOpenListAccess(), folderPath, relativeDir)
}

func (s *StorageUseCase) collectFolderZipSourcesWithAccess(access OpenListAccess, folderPath string, relativeDir string) ([]zipFileSource, error) {
	const perPage = 200

	var (
		page      uint = 1
		fetched   int
		collected []zipFileSource
	)

	for {
		data, err := s.fetchFilesRemoteWithAccess(access, folderPath, page, perPage)
		if err != nil {
			return nil, err
		}
		fetched += len(data.Content)

		var directFileItems []FileItem
		for _, item := range data.Content {
			childPath := joinStoragePath(folderPath, item.Name)
			if item.IsDir || item.Type == 1 {
				nestedDir := path.Join(relativeDir, item.Name)
				nestedSources, err := s.collectFolderZipSourcesWithAccess(access, childPath, nestedDir)
				if err != nil {
					return nil, err
				}
				collected = append(collected, nestedSources...)
				continue
			}
			directFileItems = append(directFileItems, item)
		}

		resolvedSources, err := s.resolveZipFileSourcesWithAccess(access, folderPath, relativeDir, directFileItems)
		if err != nil {
			return nil, err
		}
		collected = append(collected, resolvedSources...)

		if len(data.Content) == 0 || fetched >= data.Total || len(data.Content) < perPage {
			break
		}
		page++
	}

	return collected, nil
}

func (s *StorageUseCase) resolveZipFileSources(folderPath string, relativeDir string, items []FileItem) ([]zipFileSource, error) {
	return s.resolveZipFileSourcesWithAccess(s.defaultOpenListAccess(), folderPath, relativeDir, items)
}

func (s *StorageUseCase) resolveZipFileSourcesWithAccess(access OpenListAccess, folderPath string, relativeDir string, items []FileItem) ([]zipFileSource, error) {
	if len(items) == 0 {
		return nil, nil
	}

	const maxConcurrentResolves = 6

	results := make([]zipFileSource, len(items))
	sem := make(chan struct{}, maxConcurrentResolves)
	errCh := make(chan error, len(items))
	var wg sync.WaitGroup

	for index, item := range items {
		index := index
		item := item
		wg.Add(1)

		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			childPath := joinStoragePath(folderPath, item.Name)
			link, err := s.resolveDirectLinkWithAccess(access, childPath, childPath)
			if err != nil {
				errCh <- err
				return
			}

			results[index] = zipFileSource{
				Name:       path.Join(relativeDir, item.Name),
				DirectLink: link.DirectLink,
				Header:     link.Header,
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

func (s *StorageUseCase) fetchFileInfoRemoteWithAccess(access OpenListAccess, storagePath string) (*FileDetail, error) {
	if strings.TrimSpace(access.Token) == "" {
		return nil, fmt.Errorf("openlist token is empty")
	}
	if strings.TrimSpace(access.BaseURL) == "" {
		return nil, fmt.Errorf("openlist base url is empty")
	}

	client := http.Client{Timeout: 0}
	requestBody := map[string]interface{}{
		"path": storagePath,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", access.BaseURL+"/api/fs/get", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "OpenBridge/1.0")
	req.Header.Set("Authorization", access.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openlist /api/fs/get returned status %d", resp.StatusCode)
	}

	var fileInfoResponse FileInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&fileInfoResponse); err != nil {
		return nil, err
	}
	if fileInfoResponse.Code != http.StatusOK {
		message := strings.TrimSpace(fileInfoResponse.Message)
		if message == "" {
			message = "unknown error"
		}
		return nil, fmt.Errorf("openlist /api/fs/get failed: %s", message)
	}

	return &fileInfoResponse.Data, nil
}

func (s *StorageUseCase) streamRemoteFile(ctx context.Context, link string, headerSpec string, writer io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return err
	}

	for key, values := range parseHeaderSpec(headerSpec) {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{Timeout: 0}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("fetch remote file failed: status=%d", resp.StatusCode)
	}

	_, err = io.Copy(writer, resp.Body)
	return err
}

func parseHeaderSpec(raw string) http.Header {
	header := make(http.Header)
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return header
	}

	var stringMap map[string]string
	if err := json.Unmarshal([]byte(trimmed), &stringMap); err == nil {
		for key, value := range stringMap {
			if strings.TrimSpace(key) != "" && strings.TrimSpace(value) != "" {
				header.Add(key, value)
			}
		}
		return header
	}

	var anyMap map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &anyMap); err == nil {
		for key, value := range anyMap {
			if strings.TrimSpace(key) == "" || value == nil {
				continue
			}
			switch typed := value.(type) {
			case string:
				if strings.TrimSpace(typed) != "" {
					header.Add(key, typed)
				}
			case []interface{}:
				for _, item := range typed {
					if text := strings.TrimSpace(fmt.Sprint(item)); text != "" {
						header.Add(key, text)
					}
				}
			default:
				if text := strings.TrimSpace(fmt.Sprint(typed)); text != "" {
					header.Add(key, text)
				}
			}
		}
		return header
	}

	for _, line := range strings.Split(trimmed, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key != "" && value != "" {
			header.Add(key, value)
		}
	}

	return header
}

func joinStoragePath(parent string, name string) string {
	if parent == "" || parent == "/" {
		return "/" + strings.TrimLeft(name, "/")
	}
	return path.Clean(parent + "/" + strings.TrimLeft(name, "/"))
}

func sanitizeZipName(name string) string {
	safe := strings.ReplaceAll(strings.TrimSpace(name), "\\", "/")
	safe = path.Clean("/" + safe)
	safe = strings.TrimPrefix(safe, "/")
	safe = strings.Trim(safe, "/")
	return safe
}

func normalizeOperationNames(names []string) []string {
	result := make([]string, 0, len(names))
	seen := map[string]struct{}{}
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" || strings.Contains(trimmed, "/") || strings.Contains(trimmed, "\\") {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}
