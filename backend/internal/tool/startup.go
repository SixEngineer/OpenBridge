package tool

import (
	"fmt"
	"net/url"
	"openbridge/backend/internal/config"
	"openbridge/backend/internal/pkg/logger"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
)

func StartAria2IfNeeded(cfg config.Config, client *Aria2Client) {
	if !cfg.Aria2.AutoStart {
		return
	}

	if _, err := client.GetVersion(); err == nil {
		logger.L().Info("aria2 already running, skip autostart")
		return
	}

	aria2Path, err := resolveExecutablePath(cfg.Aria2.Path, "aria2c")
	if err != nil {
		logger.L().Warn("aria2 autostart skipped", zap.Error(err))
		return
	}

	args, err := buildAria2Args(cfg.Aria2)
	if err != nil {
		logger.L().Warn("aria2 autostart skipped", zap.Error(err))
		return
	}

	cmd := exec.Command(aria2Path, args...)
	nullFile, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0600)
	if nullFile != nil {
		defer nullFile.Close()
		cmd.Stdout = nullFile
		cmd.Stderr = nullFile
	}

	if err := cmd.Start(); err != nil {
		logger.L().Warn("aria2 autostart failed", zap.Error(err), zap.String("path", aria2Path))
		return
	}

	logger.L().Info("aria2 autostart triggered", zap.String("path", aria2Path))
}

func OpenBrowserAfterStartup(target string) {
	go func() {
		time.Sleep(700 * time.Millisecond)
		if err := openBrowser(target); err != nil {
			logger.L().Warn("open browser failed", zap.Error(err), zap.String("url", target))
		}
	}()
}

func RestartCurrentProcess(args []string) error {
	executablePath, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(executablePath, args...)
	cmd.Dir = filepath.Dir(executablePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}

func buildAria2Args(cfg config.Aria2Config) ([]string, error) {
	parsed, err := url.Parse(strings.TrimSpace(cfg.RPCURL))
	if err != nil {
		return nil, err
	}

	port := parsed.Port()
	if port == "" {
		switch parsed.Scheme {
		case "https":
			port = "443"
		default:
			port = "6800"
		}
	}

	host := strings.TrimSpace(parsed.Hostname())
	listenAll := host != "" && host != "127.0.0.1" && host != "localhost" && host != "::1"

	args := []string{
		"--enable-rpc=true",
		fmt.Sprintf("--rpc-listen-port=%s", port),
		fmt.Sprintf("--rpc-listen-all=%t", listenAll),
		"--rpc-allow-origin-all=true",
		"--continue=true",
	}

	if strings.TrimSpace(cfg.Secret) != "" {
		args = append(args, fmt.Sprintf("--rpc-secret=%s", cfg.Secret))
	}
	if strings.TrimSpace(cfg.DownloadDir) != "" {
		args = append(args, fmt.Sprintf("--dir=%s", cfg.DownloadDir))
	}

	return args, nil
}

func resolveExecutablePath(configuredPath string, fallbackName string) (string, error) {
	if strings.TrimSpace(configuredPath) != "" {
		return filepath.Clean(strings.TrimSpace(configuredPath)), nil
	}
	return exec.LookPath(fallbackName)
}

func openBrowser(target string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", target).Start()
	case "darwin":
		return exec.Command("open", target).Start()
	default:
		return exec.Command("xdg-open", target).Start()
	}
}
