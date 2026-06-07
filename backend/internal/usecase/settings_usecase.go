package usecase

import (
	"errors"
	"net/url"
	"openbridge/backend/internal/config"
	"openbridge/backend/internal/tool"
	"strings"
)

type SettingsUseCase struct {
	config      *config.Config
	aria2Client *tool.Aria2Client
}

type SettingsView struct {
	OpenListBaseURL string `json:"openlist_base_url"`
	Aria2RPCURL     string `json:"aria2_rpc_url"`
	RclonePath      string `json:"rclone_path"`
}

func NewSettingsUseCase(cfg *config.Config, aria2Client *tool.Aria2Client) *SettingsUseCase {
	return &SettingsUseCase{
		config:      cfg,
		aria2Client: aria2Client,
	}
}

func (u *SettingsUseCase) GetSettings() SettingsView {
	return SettingsView{
		OpenListBaseURL: strings.TrimSpace(u.config.OpenList.BaseURL),
		Aria2RPCURL:     strings.TrimSpace(u.config.Aria2.RPCURL),
		RclonePath:      strings.TrimSpace(u.config.Rclone.Path),
	}
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

	return u.GetSettings(), nil
}

func (u *SettingsUseCase) UpdateAria2RPCURL(rpcURL string) (SettingsView, error) {
	normalized, err := normalizeAria2RPCURL(rpcURL)
	if err != nil {
		return SettingsView{}, err
	}

	if err := config.SetEnvValues(map[string]string{
		"ARIA2_RPC_URL": normalized,
	}); err != nil {
		return SettingsView{}, err
	}

	u.config.Aria2.RPCURL = normalized
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
