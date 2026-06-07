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
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type UserUseCase struct {
	config            *config.Config
	db                *gorm.DB
	backendInstanceID string
}

func NewUserUseCase(config *config.Config, db *gorm.DB) *UserUseCase {
	return &UserUseCase{
		config:            config,
		db:                db,
		backendInstanceID: newBackendInstanceID(),
	}
}

func (uc *UserUseCase) Login(username, password string) (string, error) {

	// HTTP 客户端配置，设置超时时间为10秒
	client := &http.Client{Timeout: 10 * time.Second}

	// 构造登录请求的payload，包含用户名和密码
	payload := map[string]string{
		"username": username,
		"password": password,
	}

	// 将payload转换为JSON格式
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// 创建一个新的HTTP POST请求，目标URL为OpenList的登录接口，并将JSON数据作为请求体
	req, err := http.NewRequest("POST", normalizeSessionBaseURL(uc.config.OpenList.BaseURL)+"/api/auth/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	// 设置请求头，指定内容类型为JSON，并设置User-Agent
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "OpenBridge/1.0")

	// 发送HTTP请求并获取响应
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 解析响应体，提取登录结果
	// {
	// "code": 200,
	// "message": "success",
	// "data": {
	//     "token": "xxxx"
	// }
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

	if result.Code != 200 {
		return "", fmt.Errorf("login failed with message: %s", result.Message)
	}

	// 将登录成功后返回的Token保存到配置中，以便后续请求使用
	uc.config.OpenList.Token = result.Data.Token

	return result.Data.Token, nil
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
	// 获取所有表名
	var tables []string
	db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tables)

	// 事务清空
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
	Username          string `json:"username,omitempty"`
	Role              int    `json:"role,omitempty"`
	CheckedAt         int64  `json:"checked_at"`
	Reason            string `json:"reason,omitempty"`
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

// 获取用户数据
func (uc *UserUseCase) GetUserInfo() (UserInfo, error) {
	baseURL := normalizeSessionBaseURL(uc.config.OpenList.BaseURL)
	token := strings.TrimSpace(uc.config.OpenList.Token)
	if baseURL == "" {
		return UserInfo{}, fmt.Errorf("openlist base url is empty")
	}
	if token == "" {
		return UserInfo{}, fmt.Errorf("openlist token is empty")
	}

	// HTTP 配置，设置超时为10秒
	client := &http.Client{Timeout: 10 * time.Second}

	// 创建一个新的HTTP GET请求，目标URL为OpenList的用户信息接口
	req, err := http.NewRequest("GET", baseURL+"/api/me", nil)
	if err != nil {
		return UserInfo{}, err
	}

	// 设置请求头，指定内容类型为JSON，并设置User-Agent和Token
	req.Header.Set("User-Agent", "OpenBridge/1.0")
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	// 发送HTTP请求并获取响应
	resp, err := client.Do(req)
	if err != nil {
		return UserInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UserInfo{}, fmt.Errorf("openlist /api/me returned status %d", resp.StatusCode)
	}

	// 解析响应体，提取用户信息
	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return UserInfo{}, err
	}

	if result.Code != http.StatusOK {
		return UserInfo{}, fmt.Errorf("openlist /api/me failed: %s", result.Message)
	}

	return result.Data, nil
}

func (uc *UserUseCase) GetSessionStatus() SessionStatus {
	baseURL := normalizeSessionBaseURL(uc.config.OpenList.BaseURL)
	status := SessionStatus{
		BackendInstanceID: uc.backendInstanceID,
		OpenListBaseURL:   baseURL,
		CheckedAt:         time.Now().UnixMilli(),
	}

	switch {
	case baseURL == "":
		status.Reason = "openlist_base_url_missing"
	case strings.TrimSpace(uc.config.OpenList.Token) == "":
		status.Reason = "openlist_token_missing"
	default:
		userInfo, err := uc.GetUserInfo()
		if err != nil {
			status.Reason = "openlist_auth_invalid"
			break
		}
		status.Authenticated = true
		status.Username = userInfo.Username
		status.Role = userInfo.Role
	}

	authState := "unauthenticated"
	if status.Authenticated {
		authState = "authenticated"
	}
	status.Fingerprint = buildSessionFingerprint(
		status.BackendInstanceID,
		status.OpenListBaseURL,
		authState,
		status.Username,
		strconv.Itoa(status.Role),
		status.Reason,
	)

	return status
}

func normalizeSessionBaseURL(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
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
