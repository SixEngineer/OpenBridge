package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"openbridge/backend/internal/config"
	"openbridge/backend/internal/tool"
)

type PickLocalPathInput struct {
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	CurrentPath string `json:"current_path"`
	Filter      string `json:"filter"`
}

type PickLocalPathView struct {
	Path string `json:"path"`
}

type SystemMetricsView struct {
	CPUUsage         float64   `json:"cpu_usage"`
	MemoryUsage      float64   `json:"memory_usage"`
	MemoryUsedBytes  uint64    `json:"memory_used_bytes"`
	MemoryTotalBytes uint64    `json:"memory_total_bytes"`
	DiskUsage        float64   `json:"disk_usage"`
	DiskUsedBytes    uint64    `json:"disk_used_bytes"`
	DiskTotalBytes   uint64    `json:"disk_total_bytes"`
	DiskPath         string    `json:"disk_path"`
	Hostname         string    `json:"hostname"`
	SampledAt        time.Time `json:"sampled_at"`
}

type SystemUseCase struct {
	diskPath string
}

func NewSystemUseCase(cfg *config.Config) *SystemUseCase {
	diskPath := "."
	if cfg != nil && strings.TrimSpace(cfg.DB.Path) != "" {
		diskPath = cfg.DB.Path
	}
	return &SystemUseCase{
		diskPath: diskPath,
	}
}

func (u *SystemUseCase) PickLocalPath(input PickLocalPathInput) (PickLocalPathView, error) {
	kind := tool.PathPickerKind(strings.TrimSpace(input.Kind))
	if kind != tool.PathPickerFile && kind != tool.PathPickerDirectory {
		return PickLocalPathView{}, errors.New("path picker kind invalid")
	}

	selectedPath, err := tool.PickLocalPath(tool.PathPickerOptions{
		Kind:        kind,
		Title:       strings.TrimSpace(input.Title),
		CurrentPath: tool.NormalizeDialogPath(input.CurrentPath),
		Filter:      strings.TrimSpace(input.Filter),
	})
	if err != nil {
		return PickLocalPathView{}, err
	}

	return PickLocalPathView{
		Path: selectedPath,
	}, nil
}

func (u *SystemUseCase) GetSystemMetrics() (SystemMetricsView, error) {
	switch runtime.GOOS {
	case "windows":
		return u.getWindowsSystemMetrics()
	default:
		return SystemMetricsView{}, errors.New("system metrics are not supported on this platform yet")
	}
}

type windowsSystemMetricsSample struct {
	CPU         float64 `json:"cpu"`
	TotalMemory uint64  `json:"totalMemory"`
	FreeMemory  uint64  `json:"freeMemory"`
	TotalDisk   uint64  `json:"totalDisk"`
	FreeDisk    uint64  `json:"freeDisk"`
}

func (u *SystemUseCase) getWindowsSystemMetrics() (SystemMetricsView, error) {
	diskDrive := detectWindowsDrive(u.diskPath)
	script := strings.TrimSpace(`
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

$cpu = (Get-CimInstance Win32_Processor | Measure-Object -Property LoadPercentage -Average).Average
$os = Get-CimInstance Win32_OperatingSystem
$drive = $env:OPENBRIDGE_METRIC_DRIVE
if (-not $drive) { $drive = 'C:' }

$disk = Get-CimInstance Win32_LogicalDisk -Filter "DeviceID='$drive'"
if (-not $disk) {
  throw "logical disk not found: $drive"
}

[pscustomobject]@{
  cpu = [double]$cpu
  totalMemory = [uint64]($os.TotalVisibleMemorySize * 1024)
  freeMemory = [uint64]($os.FreePhysicalMemory * 1024)
  totalDisk = [uint64]$disk.Size
  freeDisk = [uint64]$disk.FreeSpace
} | ConvertTo-Json -Compress
`)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	cmd.Env = append(os.Environ(), "OPENBRIDGE_METRIC_DRIVE="+diskDrive)
	output, err := cmd.CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(output))
		if text == "" {
			text = err.Error()
		}
		return SystemMetricsView{}, errors.New(text)
	}

	var sample windowsSystemMetricsSample
	if err := json.Unmarshal(output, &sample); err != nil {
		return SystemMetricsView{}, err
	}

	usedMemory := safeSubtract(sample.TotalMemory, sample.FreeMemory)
	usedDisk := safeSubtract(sample.TotalDisk, sample.FreeDisk)
	hostname, _ := os.Hostname()

	return SystemMetricsView{
		CPUUsage:         clampPercent(sample.CPU),
		MemoryUsage:      usagePercent(usedMemory, sample.TotalMemory),
		MemoryUsedBytes:  usedMemory,
		MemoryTotalBytes: sample.TotalMemory,
		DiskUsage:        usagePercent(usedDisk, sample.TotalDisk),
		DiskUsedBytes:    usedDisk,
		DiskTotalBytes:   sample.TotalDisk,
		DiskPath:         diskDrive,
		Hostname:         hostname,
		SampledAt:        time.Now(),
	}, nil
}

func detectWindowsDrive(path string) string {
	if strings.TrimSpace(path) == "" {
		return "C:"
	}
	absPath, err := filepath.Abs(path)
	if err == nil {
		path = absPath
	}
	volume := filepath.VolumeName(path)
	if volume == "" {
		return "C:"
	}
	return strings.ToUpper(volume)
}

func usagePercent(used uint64, total uint64) float64 {
	if total == 0 {
		return 0
	}
	value := (float64(used) / float64(total)) * 100
	return clampPercent(value)
}

func safeSubtract(total uint64, free uint64) uint64 {
	if total <= free {
		return 0
	}
	return total - free
}

func clampPercent(value float64) float64 {
	rounded, _ := strconv.ParseFloat(fmt.Sprintf("%.1f", value), 64)
	if rounded < 0 {
		return 0
	}
	if rounded > 100 {
		return 100
	}
	return rounded
}
