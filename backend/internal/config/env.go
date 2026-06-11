package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

const envFileName = ".env"
const envExampleFileName = ".env.example"

func EnvFilePath() string {
	for _, candidate := range envCandidates(envFileName) {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	candidates := envCandidates(envFileName)
	if len(candidates) > 0 {
		return candidates[0]
	}
	return envFileName
}

func LoadEnvFile() error {
	if err := EnsureEnvFile(); err != nil {
		return err
	}
	return godotenv.Load(EnvFilePath())
}

func EnsureEnvFile() error {
	envPath := EnvFilePath()
	if _, err := os.Stat(envPath); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	examplePath := EnvExamplePath()
	if payload, err := os.ReadFile(examplePath); err == nil {
		return os.WriteFile(envPath, payload, 0644)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return WriteEnvFile(defaultEnvValues())
}

func EnvExamplePath() string {
	for _, candidate := range append(envCandidates(envExampleFileName), envCandidates(filepath.Join("backend", envExampleFileName))...) {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	candidates := envCandidates(envExampleFileName)
	if len(candidates) > 0 {
		return candidates[0]
	}
	return envExampleFileName
}

func envCandidates(name string) []string {
	candidates := make([]string, 0, 2)
	if exeDir := executableDir(); exeDir != "" {
		candidates = append(candidates, filepath.Join(exeDir, name))
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, name))
	}
	return uniquePaths(candidates)
}

func executableDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exePath)
}

func uniquePaths(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		if value == "" {
			continue
		}
		abs, err := filepath.Abs(value)
		if err == nil {
			value = abs
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func ReadEnvFile() (map[string]string, error) {
	values, err := godotenv.Read(EnvFilePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	return values, nil
}

func WriteEnvFile(values map[string]string) error {
	return godotenv.Write(values, EnvFilePath())
}

func SetEnvValues(updates map[string]string) error {
	values, err := ReadEnvFile()
	if err != nil {
		return err
	}

	for key, value := range updates {
		values[key] = value
	}

	if err := WriteEnvFile(values); err != nil {
		return err
	}

	for key, value := range updates {
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return nil
}

func GetEnvWithFallback(keys ...string) string {
	for _, key := range keys {
		value := strings.TrimSpace(os.Getenv(key))
		if value != "" {
			return value
		}
	}
	return ""
}

func NormalizeBaseURLScope(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}

func defaultEnvValues() map[string]string {
	return map[string]string{
		"APP_NAME":                 "OpenBridge",
		"APP_ENV":                  "dev",
		"APP_PORT":                 "8080",
		"APP_VERSION":              "v1.3",
		"APP_AUTO_OPEN_BROWSER":    "true",
		"DB_PATH":                  "./data/openbridge.db",
		"ARIA2_RPC_URL":            "http://127.0.0.1:6800/jsonrpc",
		"ARIA2_RPC_SECRET":         "",
		"ARIA2_DOWNLOAD_DIR":       "D:\\downloads",
		"ARIA2_PATH":               "",
		"ARIA2_AUTO_START":         "false",
		"OPENLIST_BASE_URL":        "http://127.0.0.1:5244",
		"OPENLIST_TOKEN":           "",
		"SESSION_DEVICE_LIMIT":     "5",
		"RCLONE_PATH":              "",
		"FILETREE_CACHE_MAX_BYTES": "1048576",
		"FILETREE_CACHE_DEPTH":     "2",
		"LOG_LEVEL":                "debug",
		"LOG_FORMAT":               "json",
	}
}
