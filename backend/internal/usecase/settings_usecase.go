package usecase

import (
	"errors"
	"net/url"
	"openbridge/backend/internal/config"
	"openbridge/backend/internal/tool"
	"strconv"
	"strings"
)

type SettingsUseCase struct {
	config         *config.Config
	aria2Client    *tool.Aria2Client
	storageUseCase *StorageUseCase
}

type SettingsView struct {
	OpenListBaseURL     string `json:"openlist_base_url"`
	AppVersion          string `json:"app_version"`
	Aria2RPCURL         string `json:"aria2_rpc_url"`
	Aria2Path           string `json:"aria2_path"`
	Aria2AutoStart      bool   `json:"aria2_auto_start"`
	RclonePath          string `json:"rclone_path"`
	SessionDeviceLimit  int    `json:"session_device_limit"`
	AutoOpenBrowser     bool   `json:"auto_open_browser"`
	FileTreeCacheSizeKB int64  `json:"filetree_cache_size_kb"`
	FileTreeCacheDepth  int    `json:"filetree_cache_depth"`
}

func NewSettingsUseCase(cfg *config.Config, aria2Client *tool.Aria2Client, storageUseCase *StorageUseCase) *SettingsUseCase {
	return &SettingsUseCase{
		config:         cfg,
		aria2Client:    aria2Client,
		storageUseCase: storageUseCase,
	}
}

func (u *SettingsUseCase) GetSettings() SettingsView {
	cacheSizeKB := u.config.FileTree.MaxBytes / 1024
	if cacheSizeKB <= 0 {
		cacheSizeKB = 4
	}

	return SettingsView{
		OpenListBaseURL:     strings.TrimSpace(u.config.OpenList.BaseURL),
		AppVersion:          normalizeAppVersion(u.config.App.Version),
		Aria2RPCURL:         strings.TrimSpace(u.config.Aria2.RPCURL),
		Aria2Path:           strings.TrimSpace(u.config.Aria2.Path),
		Aria2AutoStart:      u.config.Aria2.AutoStart,
		RclonePath:          strings.TrimSpace(u.config.Rclone.Path),
		SessionDeviceLimit:  u.config.Auth.SessionDeviceLimit,
		AutoOpenBrowser:     u.config.App.AutoOpenBrowser,
		FileTreeCacheSizeKB: cacheSizeKB,
		FileTreeCacheDepth:  u.config.FileTree.MaxDepth,
	}
}

func normalizeAppVersion(version string) string {
	trimmed := strings.TrimSpace(version)
	if trimmed == "" {
		return "v1.2"
	}
	return trimmed
}

func (u *SettingsUseCase) UpdateOpenListBaseURL(baseURL string) (SettingsView, error) {
	normalized, err := normalizeOpenListBaseURL(baseURL)
	if err != nil {
		return SettingsView{}, err
	}

	if err := config.SetEnvValues(map[string]string{
		"OPENLIST_BASE_URL": normalized,
	}); err != nil {
		return SettingsView{}, err
	}

	u.config.OpenList.BaseURL = normalized
	if u.storageUseCase != nil {
		if err := u.storageUseCase.UpdateOpenListSource(); err != nil {
			return SettingsView{}, err
		}
	}

	return u.GetSettings(), nil
}

func (u *SettingsUseCase) UpdateAria2(rpcURL, aria2Path string, autoStart bool) (SettingsView, error) {
	trimmedRPCURL := strings.TrimSpace(rpcURL)
	if trimmedRPCURL == "" {
		trimmedRPCURL = strings.TrimSpace(u.config.Aria2.RPCURL)
	}
	if trimmedRPCURL == "" {
		trimmedRPCURL = "http://127.0.0.1:6800/jsonrpc"
	}

	normalized, err := normalizeAria2RPCURL(trimmedRPCURL)
	if err != nil {
		return SettingsView{}, err
	}

	normalizedPath, err := normalizeOptionalExecutablePath(aria2Path)
	if err != nil {
		return SettingsView{}, err
	}

	if err := config.SetEnvValues(map[string]string{
		"ARIA2_RPC_URL":    normalized,
		"ARIA2_PATH":       normalizedPath,
		"ARIA2_AUTO_START": strconv.FormatBool(autoStart),
	}); err != nil {
		return SettingsView{}, err
	}

	u.config.Aria2.RPCURL = normalized
	u.config.Aria2.Path = normalizedPath
	u.config.Aria2.AutoStart = autoStart
	u.aria2Client.SetConfig(normalized, u.config.Aria2.Secret)

	return u.GetSettings(), nil
}

func (u *SettingsUseCase) UpdateRclonePath(rclonePath string) (SettingsView, error) {
	normalized, err := normalizeRclonePath(rclonePath)
	if err != nil {
		return SettingsView{}, err
	}

	if err := config.SetEnvValues(map[string]string{
		"RCLONE_PATH": normalized,
	}); err != nil {
		return SettingsView{}, err
	}

	u.config.Rclone.Path = normalized
	return u.GetSettings(), nil
}

func (u *SettingsUseCase) UpdateSessionDeviceLimit(limit int) (SettingsView, error) {
	normalized, err := normalizeSessionDeviceLimit(limit)
	if err != nil {
		return SettingsView{}, err
	}

	if err := config.SetEnvValues(map[string]string{
		"SESSION_DEVICE_LIMIT": normalized,
	}); err != nil {
		return SettingsView{}, err
	}

	u.config.Auth.SessionDeviceLimit = limit
	return u.GetSettings(), nil
}

func (u *SettingsUseCase) UpdateApp(autoOpenBrowser bool) (SettingsView, error) {
	if err := config.SetEnvValues(map[string]string{
		"APP_AUTO_OPEN_BROWSER": strconv.FormatBool(autoOpenBrowser),
	}); err != nil {
		return SettingsView{}, err
	}

	u.config.App.AutoOpenBrowser = autoOpenBrowser
	return u.GetSettings(), nil
}

func (u *SettingsUseCase) UpdateFileTree(cacheSizeKB int64, cacheDepth int) (SettingsView, error) {
	normalizedSizeKB, err := normalizeFileTreeCacheSizeKB(cacheSizeKB)
	if err != nil {
		return SettingsView{}, err
	}

	normalizedDepth, err := normalizeFileTreeCacheDepth(cacheDepth)
	if err != nil {
		return SettingsView{}, err
	}

	maxBytes := normalizedSizeKB * 1024
	if err := config.SetEnvValues(map[string]string{
		"FILETREE_CACHE_MAX_BYTES": strconv.FormatInt(maxBytes, 10),
		"FILETREE_CACHE_DEPTH":     strconv.Itoa(normalizedDepth),
	}); err != nil {
		return SettingsView{}, err
	}

	u.config.FileTree.MaxBytes = maxBytes
	u.config.FileTree.MaxDepth = normalizedDepth
	if u.storageUseCase != nil {
		if err := u.storageUseCase.UpdateFileTreeConfig(maxBytes, normalizedDepth); err != nil {
			return SettingsView{}, err
		}
	}

	return u.GetSettings(), nil
}

func normalizeOpenListBaseURL(baseURL string) (string, error) {
	normalized := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if normalized == "" {
		return "", errors.New("openlist base url empty")
	}

	parsed, err := url.Parse(normalized)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("openlist base url invalid")
	}

	return normalized, nil
}

func normalizeAria2RPCURL(rpcURL string) (string, error) {
	normalized := strings.TrimSpace(rpcURL)
	if normalized == "" {
		return "", errors.New("aria2 rpc url empty")
	}

	parsed, err := url.Parse(normalized)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("aria2 rpc url invalid")
	}

	return normalized, nil
}

func normalizeRclonePath(rclonePath string) (string, error) {
	normalized := strings.TrimSpace(rclonePath)
	if normalized == "" {
		return "", errors.New("rclone path empty")
	}
	return normalized, nil
}

func normalizeOptionalExecutablePath(value string) (string, error) {
	return strings.TrimSpace(value), nil
}

func normalizeSessionDeviceLimit(limit int) (string, error) {
	if limit < 1 {
		return "", errors.New("session device limit invalid")
	}
	return strconv.Itoa(limit), nil
}

func normalizeFileTreeCacheSizeKB(sizeKB int64) (int64, error) {
	if sizeKB < 4 {
		return 0, errors.New("filetree cache size must be at least 4KB")
	}
	return sizeKB, nil
}

func normalizeFileTreeCacheDepth(depth int) (int, error) {
	if depth < 1 || depth > 5 {
		return 0, errors.New("filetree cache depth must be between 1 and 5")
	}
	return depth, nil
}
