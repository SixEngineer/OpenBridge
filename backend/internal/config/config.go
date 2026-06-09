package config

import (
	"os"
	"strconv"
)

type Config struct {
	App      AppConfig
	DB       DBConfig
	Aria2    Aria2Config
	OpenList OpenListConfig
	Auth     AuthConfig
	Rclone   RcloneConfig
	FileTree FileTreeConfig
	Log      LogConfig
}

type AppConfig struct {
	Name            string
	Env             string
	Port            string
	Version         string
	AutoOpenBrowser bool
}

type DBConfig struct {
	Path string
}

type Aria2Config struct {
	RPCURL      string
	Secret      string
	DownloadDir string
	Path        string
	AutoStart   bool
}

type OpenListConfig struct {
	BaseURL string
	Token   string
}

type AuthConfig struct {
	SessionDeviceLimit int
}

type RcloneConfig struct {
	Path string
}

type FileTreeConfig struct {
	MaxBytes int64
	MaxDepth int
}

type LogConfig struct {
	Level  string
	Format string
}

// 读取配置
func ReadConfig() Config {
	// 加载.env文件
	err := LoadEnvFile()
	if err != nil {
		panic("Error loading .env file")
	}

	// 从环境变量中读取配置
	return Config{
		App: AppConfig{
			Name:            os.Getenv("APP_NAME"),
			Env:             os.Getenv("APP_ENV"),
			Port:            os.Getenv("APP_PORT"),
			Version:         os.Getenv("APP_VERSION"),
			AutoOpenBrowser: getEnvBool("APP_AUTO_OPEN_BROWSER", true),
		},
		DB: DBConfig{
			Path: os.Getenv("DB_PATH"),
		},
		Aria2: Aria2Config{
			RPCURL:      os.Getenv("ARIA2_RPC_URL"),
			Secret:      GetEnvWithFallback("ARIA2_SECRET", "ARIA2_RPC_SECRET"),
			DownloadDir: os.Getenv("ARIA2_DOWNLOAD_DIR"),
			Path:        GetEnvWithFallback("ARIA2_PATH", "ARIA2C_PATH"),
			AutoStart:   getEnvBool("ARIA2_AUTO_START", false),
		},
		OpenList: OpenListConfig{
			BaseURL: os.Getenv("OPENLIST_BASE_URL"),
			Token:   os.Getenv("OPENLIST_TOKEN"),
		},
		Auth: AuthConfig{
			SessionDeviceLimit: getEnvInt("SESSION_DEVICE_LIMIT", 5),
		},
		Rclone: RcloneConfig{
			Path: os.Getenv("RCLONE_PATH"),
		},
		FileTree: FileTreeConfig{
			MaxBytes: getEnvInt64("FILETREE_CACHE_MAX_BYTES", 1024*1024, 4*1024),
			MaxDepth: getEnvRangedInt("FILETREE_CACHE_DEPTH", 2, 1, 5),
		},
		Log: LogConfig{
			Level:  os.Getenv("LOG_LEVEL"),
			Format: os.Getenv("LOG_FORMAT"),
		},
	}
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

func getEnvInt64(key string, fallback int64, minimum int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < minimum {
		return fallback
	}

	return parsed
}

func getEnvRangedInt(key string, fallback int, minimum int, maximum int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < minimum || parsed > maximum {
		return fallback
	}

	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	switch value {
	case "1", "true", "TRUE", "True", "yes", "YES", "on", "ON":
		return true
	case "0", "false", "FALSE", "False", "no", "NO", "off", "OFF":
		return false
	default:
		return fallback
	}
}
