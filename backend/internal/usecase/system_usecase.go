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
	"sync"
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
	CPUUsage           float64   `json:"cpu_usage"`
	ProcessCPUUsage    float64   `json:"process_cpu_usage"`
	MemoryUsage        float64   `json:"memory_usage"`
	MemoryUsedBytes    uint64    `json:"memory_used_bytes"`
	MemoryTotalBytes   uint64    `json:"memory_total_bytes"`
	ProcessMemoryBytes uint64    `json:"process_memory_bytes"`
	DiskUsage          float64   `json:"disk_usage"`
	DiskUsedBytes      uint64    `json:"disk_used_bytes"`
	DiskTotalBytes     uint64    `json:"disk_total_bytes"`
	AppDiskUsageBytes  uint64    `json:"app_disk_usage_bytes"`
	NetworkReceiveBPS  uint64    `json:"network_receive_bytes_per_sec"`
	NetworkTransmitBPS uint64    `json:"network_transmit_bytes_per_sec"`
	DiskPath           string    `json:"disk_path"`
	Hostname           string    `json:"hostname"`
	SampledAt          time.Time `json:"sampled_at"`
}

type SystemUseCase struct {
	diskPath          string
	appDiskPaths      []string
	runtimeController systemRuntimeController
	metricsMu         sync.Mutex
	lastNetworkSample *networkRateSample
}

type systemRuntimeController interface {
	RequestExit()
	RequestRestart()
}

type ServiceControlView struct {
	Accepted bool   `json:"accepted"`
	Action   string `json:"action"`
}

func NewSystemUseCase(cfg *config.Config, runtimeController systemRuntimeController) *SystemUseCase {
	diskPath := "."
	if cfg != nil && strings.TrimSpace(cfg.DB.Path) != "" {
		diskPath = cfg.DB.Path
	}
	return &SystemUseCase{
		diskPath:          diskPath,
		appDiskPaths:      collectAppDiskPaths(cfg),
		runtimeController: runtimeController,
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

func (u *SystemUseCase) RestartApplication() ServiceControlView {
	if u.runtimeController != nil {
		go func() {
			time.Sleep(200 * time.Millisecond)
			u.runtimeController.RequestRestart()
		}()
	}
	return ServiceControlView{Accepted: true, Action: "restart"}
}

func (u *SystemUseCase) ExitApplication() ServiceControlView {
	if u.runtimeController != nil {
		go func() {
			time.Sleep(200 * time.Millisecond)
			u.runtimeController.RequestExit()
		}()
	}
	return ServiceControlView{Accepted: true, Action: "exit"}
}

type windowsSystemMetricsSample struct {
	CPU           float64 `json:"cpu"`
	ProcessCPU    float64 `json:"processCpu"`
	TotalMemory   uint64  `json:"totalMemory"`
	FreeMemory    uint64  `json:"freeMemory"`
	ProcessMemory uint64  `json:"processMemory"`
	TotalDisk     uint64  `json:"totalDisk"`
	FreeDisk      uint64  `json:"freeDisk"`
	ReceivedBytes uint64  `json:"receivedBytes"`
	SentBytes     uint64  `json:"sentBytes"`
}

type networkRateSample struct {
	receivedBytes uint64
	sentBytes     uint64
	recordedAt    time.Time
}

func (u *SystemUseCase) getWindowsSystemMetrics() (SystemMetricsView, error) {
	diskDrive := detectWindowsDrive(u.diskPath)
	script := strings.TrimSpace(`
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

$cpu = (Get-CimInstance Win32_Processor | Measure-Object -Property LoadPercentage -Average).Average
$os = Get-CimInstance Win32_OperatingSystem
$computer = Get-CimInstance Win32_ComputerSystem
$drive = $env:OPENBRIDGE_METRIC_DRIVE
if (-not $drive) { $drive = 'C:' }
$pidValue = $env:OPENBRIDGE_METRIC_PID
$processCpu = 0
$processMemory = 0
$receivedBytes = 0
$sentBytes = 0

$disk = Get-CimInstance Win32_LogicalDisk -Filter "DeviceID='$drive'"
if (-not $disk) {
  throw "logical disk not found: $drive"
}

if ($pidValue) {
  $procPerf = Get-CimInstance Win32_PerfFormattedData_PerfProc_Process | Where-Object { $_.IDProcess -eq [int]$pidValue } | Select-Object -First 1
  if ($procPerf) {
    $logicalCpu = [double]$computer.NumberOfLogicalProcessors
    if ($logicalCpu -gt 0) {
      $processCpu = [double]$procPerf.PercentProcessorTime / $logicalCpu
    } else {
      $processCpu = [double]$procPerf.PercentProcessorTime
    }
    $processMemory = [uint64]$procPerf.WorkingSetPrivate
  }
}

$adapterStats = Get-NetAdapterStatistics -ErrorAction SilentlyContinue | Where-Object {
  $_.ReceivedBytes -ne $null -and $_.SentBytes -ne $null
}
if ($adapterStats) {
  foreach ($adapter in $adapterStats) {
    $receivedBytes += [uint64]$adapter.ReceivedBytes
    $sentBytes += [uint64]$adapter.SentBytes
  }
}

[pscustomobject]@{
  cpu = [double]$cpu
  processCpu = [double]$processCpu
  totalMemory = [uint64]($os.TotalVisibleMemorySize * 1024)
  freeMemory = [uint64]($os.FreePhysicalMemory * 1024)
  processMemory = [uint64]$processMemory
  totalDisk = [uint64]$disk.Size
  freeDisk = [uint64]$disk.FreeSpace
  receivedBytes = [uint64]$receivedBytes
  sentBytes = [uint64]$sentBytes
} | ConvertTo-Json -Compress
`)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	cmd.Env = append(
		os.Environ(),
		"OPENBRIDGE_METRIC_DRIVE="+diskDrive,
		"OPENBRIDGE_METRIC_PID="+strconv.Itoa(os.Getpid()),
	)
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
	appDiskUsage := sumPathSizes(u.appDiskPaths)
	hostname, _ := os.Hostname()
	sampledAt := time.Now()
	receiveRate, transmitRate := u.measureNetworkRates(sample.ReceivedBytes, sample.SentBytes, sampledAt)

	return SystemMetricsView{
		CPUUsage:           clampPercent(sample.CPU),
		ProcessCPUUsage:    clampPercent(sample.ProcessCPU),
		MemoryUsage:        usagePercent(usedMemory, sample.TotalMemory),
		MemoryUsedBytes:    usedMemory,
		MemoryTotalBytes:   sample.TotalMemory,
		ProcessMemoryBytes: sample.ProcessMemory,
		DiskUsage:          usagePercent(usedDisk, sample.TotalDisk),
		DiskUsedBytes:      usedDisk,
		DiskTotalBytes:     sample.TotalDisk,
		AppDiskUsageBytes:  appDiskUsage,
		NetworkReceiveBPS:  receiveRate,
		NetworkTransmitBPS: transmitRate,
		DiskPath:           diskDrive,
		Hostname:           hostname,
		SampledAt:          sampledAt,
	}, nil
}

func (u *SystemUseCase) measureNetworkRates(receivedBytes uint64, sentBytes uint64, recordedAt time.Time) (uint64, uint64) {
	u.metricsMu.Lock()
	defer u.metricsMu.Unlock()

	current := &networkRateSample{
		receivedBytes: receivedBytes,
		sentBytes:     sentBytes,
		recordedAt:    recordedAt,
	}

	previous := u.lastNetworkSample
	u.lastNetworkSample = current
	if previous == nil {
		return 0, 0
	}

	elapsed := recordedAt.Sub(previous.recordedAt).Seconds()
	if elapsed <= 0 {
		return 0, 0
	}

	receivedDelta := safeSubtract(receivedBytes, previous.receivedBytes)
	sentDelta := safeSubtract(sentBytes, previous.sentBytes)

	return uint64(float64(receivedDelta) / elapsed), uint64(float64(sentDelta) / elapsed)
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

func collectAppDiskPaths(cfg *config.Config) []string {
	paths := make([]string, 0, 3)
	if cfg != nil && strings.TrimSpace(cfg.DB.Path) != "" {
		paths = append(paths, strings.TrimSpace(cfg.DB.Path))
	}
	if exePath, err := os.Executable(); err == nil && strings.TrimSpace(exePath) != "" {
		paths = append(paths, exePath)
	}
	return uniqueStringPaths(paths)
}

func uniqueStringPaths(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		normalized := value
		if absValue, err := filepath.Abs(value); err == nil {
			normalized = absValue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func sumPathSizes(paths []string) uint64 {
	var total uint64
	for _, path := range uniqueStringPaths(paths) {
		total += pathSize(path)
	}
	return total
}

func pathSize(path string) uint64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	if !info.IsDir() {
		return uint64(info.Size())
	}

	var total uint64
	_ = filepath.Walk(path, func(_ string, fileInfo os.FileInfo, walkErr error) error {
		if walkErr != nil || fileInfo == nil || fileInfo.IsDir() {
			return nil
		}
		total += uint64(fileInfo.Size())
		return nil
	})
	return total
}
