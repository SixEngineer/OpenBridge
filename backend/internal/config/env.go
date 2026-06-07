package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

const envFileName = ".env"

func EnvFilePath() string {
	path, err := filepath.Abs(envFileName)
	if err != nil {
		return envFileName
	}
	return path
}

func LoadEnvFile() error {
	return godotenv.Load(EnvFilePath())
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
