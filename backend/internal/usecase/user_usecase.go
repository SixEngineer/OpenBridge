package usecase

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"openbridge/backend/internal/config"
	"openbridge/backend/internal/domain/entity"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

const (
	DeviceIDHeader       = "X-OpenBridge-Device-ID"
	legacyDeviceID       = "legacy-browser"
	deviceSessionMaxIdle = 30 * 24 * time.Hour
)

type UserUseCase struct {
	config            *config.Config
	db                *gorm.DB
	backendInstanceID string
	sessionMu         sync.RWMutex
	sessions          map[string]*deviceSession
}

type deviceSession struct {
	DeviceID        string
	Username        string
	Role            int
	Token           string
	IssuedAt        int64
	LastSeenAt      int64
	OpenListBaseURL string
}

func NewUserUseCase(config *config.Config, db *gorm.DB) *UserUseCase {
	return &UserUseCase{
		config:            config,
		db:                db,
		backendInstanceID: newBackendInstanceID(),
		sessions:          make(map[string]*deviceSession),
	}
}

func (uc *UserUseCase) Login(username, password, deviceID string) (string, error) {
	baseURL := normalizeSessionBaseURL(uc.config.OpenList.BaseURL)
	if baseURL == "" {
		return "", errors.New("openlist base url is empty")
	}

	token, err := uc.loginOpenList(baseURL, username, password)
	if err != nil {
		return "", err
	}

	userInfo, err := uc.fetchUserInfoByToken(baseURL, token)
	if err != nil {
		return "", err
	}

	normalizedDeviceID := normalizeDeviceID(deviceID)
	now := time.Now().UnixMilli()
	limit := uc.getSessionDeviceLimit()

	uc.sessionMu.Lock()
	defer uc.sessionMu.Unlock()

	uc.pruneSessionsLocked(baseURL, now)
	if uc.countUserSessionsLocked(userInfo.Username, baseURL, normalizedDeviceID) >= limit {
		return "", fmt.Errorf("device_limit_reached: at most %d devices can stay logged in at the same time", limit)
	}

	uc.sessions[normalizedDeviceID] = &deviceSession{
		DeviceID:        normalizedDeviceID,
		Username:        userInfo.Username,
		Role:            userInfo.Role,
		Token:           token,
		IssuedAt:        now,
		LastSeenAt:      now,
		OpenListBaseURL: baseURL,
	}

	// 保留全局 token，兼容当前仍依赖配置 token 的后端逻辑。
	uc.config.OpenList.Token = token

	return token, nil
}

func (uc *UserUseCase) loginOpenList(baseURL, username, password string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	payload := map[string]string{
		"username": username,
		"password": password,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", baseURL+"/api/auth/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "OpenBridge/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Code != http.StatusOK {
		return "", fmt.Errorf("login failed with message: %s", result.Message)
	}

	return strings.TrimSpace(result.Data.Token), nil
}

type ResetScope string

const (
	ResetScopeCurrent ResetScope = "current"
	ResetScopeAll     ResetScope = "all"
)

func (uc *UserUseCase) Reset(scope ResetScope) error {
	switch scope {
	case ResetScopeCurrent:
		return uc.clearCurrentOpenListData()
	case ResetScopeAll:
		return uc.ClearAllTables(uc.db)
	default:
		return errors.New("invalid reset scope")
	}
}

func (uc *UserUseCase) ClearAllTables(db *gorm.DB) error {
	var tables []string
	db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tables)

	return db.Transaction(func(tx *gorm.DB) error {
		tx.Exec("PRAGMA foreign_keys = OFF")
		for _, table := range tables {
			tx.Exec(fmt.Sprintf("DELETE FROM %s", table))
		}
		tx.Exec("PRAGMA foreign_keys = ON")
		tx.Exec("VACUUM")
		return nil
	})
}

func (uc *UserUseCase) clearCurrentOpenListData() error {
	scope := config.NormalizeBaseURLScope(uc.config.OpenList.BaseURL)
	if scope == "" {
		return errors.New("openlist base url is empty")
	}

	return uc.db.Transaction(func(tx *gorm.DB) error {
		var providerIDs []uint
		if err := tx.Model(&entity.ProviderAccount{}).
			Where("openlist_base_url = ?", scope).
			Pluck("id", &providerIDs).Error; err != nil {
			return err
		}

		var mountIDs []uint
		if len(providerIDs) > 0 {
			if err := tx.Model(&entity.MountPoint{}).
				Where("provider_account_id IN ?", providerIDs).
				Pluck("id", &mountIDs).Error; err != nil {
				return err
			}
		}

		if len(mountIDs) > 0 {
			if err := tx.Where("inherit_from_id IN ?", mountIDs).Delete(&entity.MountPoint{}).Error; err != nil {
				return err
			}
		}

		if len(providerIDs) > 0 {
			if err := tx.Where("provider_account_id IN ?", providerIDs).Delete(&entity.QuotaSnapshot{}).Error; err != nil {
				return err
			}
			if err := tx.Where("provider_account_id IN ?", providerIDs).Delete(&entity.MountPoint{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", providerIDs).Delete(&entity.ProviderAccount{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("openlist_base_url = ?", scope).Delete(&entity.DownloadTask{}).Error; err != nil {
			return err
		}

		return nil
	})
}

type Response struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    UserInfo `json:"data"`
}

type SessionStatus struct {
	Authenticated     bool   `json:"authenticated"`
	BackendInstanceID string `json:"backend_instance_id"`
	OpenListBaseURL   string `json:"openlist_base_url"`
	Fingerprint       string `json:"fingerprint"`
	DeviceID          string `json:"device_id"`
	DeviceLimit       int    `json:"device_limit"`
	ActiveDeviceCount int    `json:"active_device_count"`
	Username          string `json:"username,omitempty"`
	Role              int    `json:"role,omitempty"`
	CheckedAt         int64  `json:"checked_at"`
	Reason            string `json:"reason,omitempty"`
}

type OpenListAccess struct {
	BaseURL  string
	Token    string
	Username string
	BasePath string
	Role     int
}

type UserInfo struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	BasePath   string `json:"base_path"`
	Role       int    `json:"role"`
	Disabled   bool   `json:"disabled"`
	Permission int    `json:"permission"`
	SSOID      string `json:"sso_id"`
	OTP        bool   `json:"otp"`
}

func (uc *UserUseCase) GetUserInfo(deviceID string) (UserInfo, error) {
	baseURL := normalizeSessionBaseURL(uc.config.OpenList.BaseURL)
	if baseURL == "" {
		return UserInfo{}, fmt.Errorf("openlist base url is empty")
	}

	session := uc.getSession(deviceID)
	if session == nil {
		return UserInfo{}, fmt.Errorf("device session missing")
	}
	if session.OpenListBaseURL != baseURL {
		uc.removeSession(deviceID)
		return UserInfo{}, fmt.Errorf("openlist source changed")
	}

	token := uc.getTokenForDevice(deviceID)
	if token == "" {
		return UserInfo{}, fmt.Errorf("openlist token is empty")
	}

	userInfo, err := uc.fetchUserInfoByToken(baseURL, token)
	if err != nil {
		uc.removeSession(normalizeDeviceID(deviceID))
		return UserInfo{}, err
	}

	uc.touchSession(normalizeDeviceID(deviceID), userInfo, token, baseURL)
	return userInfo, nil
}

func (uc *UserUseCase) GetOpenListAccess(deviceID string) (OpenListAccess, error) {
	userInfo, err := uc.GetUserInfo(deviceID)
	if err != nil {
		return OpenListAccess{}, err
	}

	session := uc.getSession(deviceID)
	if session == nil {
		return OpenListAccess{}, fmt.Errorf("device session missing")
	}

	token := strings.TrimSpace(session.Token)
	if token == "" {
		return OpenListAccess{}, fmt.Errorf("openlist token is empty")
	}

	return OpenListAccess{
		BaseURL:  normalizeSessionBaseURL(uc.config.OpenList.BaseURL),
		Token:    token,
		Username: userInfo.Username,
		BasePath: userInfo.BasePath,
		Role:     userInfo.Role,
	}, nil
}

func (uc *UserUseCase) fetchUserInfoByToken(baseURL, token string) (UserInfo, error) {
	if strings.TrimSpace(token) == "" {
		return UserInfo{}, fmt.Errorf("openlist token is empty")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", baseURL+"/api/me", nil)
	if err != nil {
		return UserInfo{}, err
	}

	req.Header.Set("User-Agent", "OpenBridge/1.0")
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return UserInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UserInfo{}, fmt.Errorf("openlist /api/me returned status %d", resp.StatusCode)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return UserInfo{}, err
	}

	if result.Code != http.StatusOK {
		return UserInfo{}, fmt.Errorf("openlist /api/me failed: %s", result.Message)
	}

	return result.Data, nil
}

func (uc *UserUseCase) GetSessionStatus(deviceID string) SessionStatus {
	baseURL := normalizeSessionBaseURL(uc.config.OpenList.BaseURL)
	normalizedDeviceID := normalizeDeviceID(deviceID)
	status := SessionStatus{
		BackendInstanceID: uc.backendInstanceID,
		OpenListBaseURL:   baseURL,
		DeviceID:          normalizedDeviceID,
		DeviceLimit:       uc.getSessionDeviceLimit(),
		CheckedAt:         time.Now().UnixMilli(),
	}

	switch {
	case baseURL == "":
		status.Reason = "openlist_base_url_missing"
	case uc.getSession(normalizedDeviceID) == nil:
		status.Reason = "device_session_missing"
	case uc.getSession(normalizedDeviceID).OpenListBaseURL != baseURL:
		uc.removeSession(normalizedDeviceID)
		status.Reason = "source_changed"
	default:
		userInfo, err := uc.GetUserInfo(normalizedDeviceID)
		if err != nil {
			if strings.Contains(err.Error(), "source changed") {
				status.Reason = "source_changed"
			} else {
				status.Reason = "openlist_auth_invalid"
			}
			break
		}
		status.Authenticated = true
		status.Username = userInfo.Username
		status.Role = userInfo.Role
		status.ActiveDeviceCount = uc.countUserSessions(userInfo.Username, baseURL)
	}

	status.Fingerprint = buildSessionFingerprint(
		status.BackendInstanceID,
		status.OpenListBaseURL,
		status.DeviceID,
		status.Username,
	)

	return status
}

func (uc *UserUseCase) getTokenForDevice(deviceID string) string {
	if session := uc.getSession(deviceID); session != nil {
		return strings.TrimSpace(session.Token)
	}
	return ""
}

func (uc *UserUseCase) touchSession(deviceID string, userInfo UserInfo, token, baseURL string) {
	normalizedDeviceID := normalizeDeviceID(deviceID)
	now := time.Now().UnixMilli()

	uc.sessionMu.Lock()
	defer uc.sessionMu.Unlock()

	session := uc.sessions[normalizedDeviceID]
	if session == nil {
		uc.sessions[normalizedDeviceID] = &deviceSession{
			DeviceID:        normalizedDeviceID,
			Username:        userInfo.Username,
			Role:            userInfo.Role,
			Token:           token,
			IssuedAt:        now,
			LastSeenAt:      now,
			OpenListBaseURL: baseURL,
		}
		return
	}

	session.Username = userInfo.Username
	session.Role = userInfo.Role
	session.Token = token
	session.LastSeenAt = now
	session.OpenListBaseURL = baseURL
}

func (uc *UserUseCase) removeSession(deviceID string) {
	normalizedDeviceID := normalizeDeviceID(deviceID)
	uc.sessionMu.Lock()
	defer uc.sessionMu.Unlock()
	delete(uc.sessions, normalizedDeviceID)
}

func (uc *UserUseCase) getSession(deviceID string) *deviceSession {
	normalizedDeviceID := normalizeDeviceID(deviceID)
	uc.sessionMu.RLock()
	defer uc.sessionMu.RUnlock()
	return uc.sessions[normalizedDeviceID]
}

func (uc *UserUseCase) countUserSessions(username, baseURL string) int {
	uc.sessionMu.RLock()
	defer uc.sessionMu.RUnlock()
	return uc.countUserSessionsLocked(username, baseURL, "")
}

func (uc *UserUseCase) countUserSessionsLocked(username, baseURL, excludeDeviceID string) int {
	count := 0
	for deviceID, session := range uc.sessions {
		if excludeDeviceID != "" && deviceID == excludeDeviceID {
			continue
		}
		if session.Username == username && session.OpenListBaseURL == baseURL {
			count++
		}
	}
	return count
}

func (uc *UserUseCase) pruneSessionsLocked(baseURL string, now int64) {
	maxIdleMillis := deviceSessionMaxIdle.Milliseconds()
	for deviceID, session := range uc.sessions {
		switch {
		case session == nil:
			delete(uc.sessions, deviceID)
		case session.OpenListBaseURL != baseURL:
			delete(uc.sessions, deviceID)
		case strings.TrimSpace(session.Token) == "":
			delete(uc.sessions, deviceID)
		case now-session.LastSeenAt > maxIdleMillis:
			delete(uc.sessions, deviceID)
		}
	}
}

func (uc *UserUseCase) getSessionDeviceLimit() int {
	if uc.config.Auth.SessionDeviceLimit <= 0 {
		return 5
	}
	return uc.config.Auth.SessionDeviceLimit
}

func normalizeSessionBaseURL(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}

func normalizeDeviceID(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return legacyDeviceID
	}
	return normalized
}

func newBackendInstanceID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

func buildSessionFingerprint(parts ...string) string {
	hash := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(hash[:])
}
