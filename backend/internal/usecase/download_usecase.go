package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"openbridge/backend/internal/config"
	"openbridge/backend/internal/domain/entity"
	"openbridge/backend/internal/repository"
	"openbridge/backend/internal/tool"
)

type DownloadUseCase struct {
	storageUseCase *StorageUseCase
	downloadRepo   *repository.DownloadRepository
	aria2Client    *tool.Aria2Client
	config         *config.Config
}

func NewDownloadUseCase(storageUseCase *StorageUseCase, downloadRepo *repository.DownloadRepository, aria2Client *tool.Aria2Client, config *config.Config) *DownloadUseCase {
	return &DownloadUseCase{
		storageUseCase: storageUseCase,
		downloadRepo:   downloadRepo,
		aria2Client:    aria2Client,
		config:         config,
	}
}

func (u *DownloadUseCase) CreateTask(deviceID string, path string, dir string) (*entity.DownloadTask, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("path empty")
	}

	directLink, err := u.storageUseCase.ResolveDirectLinkForDevice(deviceID, path)
	if err != nil {
		return nil, err
	}

	options := map[string]interface{}{}
	targetDir := strings.TrimSpace(dir)
	if targetDir == "" {
		targetDir = strings.TrimSpace(u.config.Aria2.DownloadDir)
	}
	if targetDir != "" {
		options["dir"] = targetDir
	}

	gid := ""
	if len(options) == 0 {
		gid, err = u.aria2Client.AddURI(directLink.DirectLink)
	} else {
		gid, err = u.aria2Client.AddURIWithOptions(directLink.DirectLink, options)
	}
	if err != nil {
		return nil, err
	}

	task := &entity.DownloadTask{
		TaskID:          newTaskID(),
		OpenListBaseURL: config.NormalizeBaseURLScope(u.config.OpenList.BaseURL),
		SourcePath:      directLink.StoragePath,
		DirectLink:      directLink.DirectLink,
		FileName:        directLink.Name,
		FileSize:        directLink.Size,
		Aria2GID:        gid,
		Status:          "waiting",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := u.downloadRepo.InsertTask(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (u *DownloadUseCase) GetTask(taskID string) (*entity.DownloadTask, error) {
	if strings.TrimSpace(taskID) == "" {
		return nil, errors.New("task_id empty")
	}

	task, err := u.downloadRepo.GetTaskByTaskID(taskID)
	if err != nil {
		return nil, err
	}
	if task.Status == "deleted" {
		return task, nil
	}

	if status, err := u.aria2Client.TellStatus(task.Aria2GID); err == nil {
		u.applyAria2Status(task, status)
		u.downloadRepo.UpdateTask(task)
	} else if strings.Contains(err.Error(), "not found") && task.Status != "complete" {
		// GID not found in aria2 -> task was removed or expired -> mark as error.
		if task.Status != "stopped" {
			task.Status = "error"
		}
		u.downloadRepo.UpdateTask(task)
	} // Network errors (aria2 unreachable) or already complete: keep current DB status.

	return task, nil
}

func (u *DownloadUseCase) StopTask(taskID string) (*entity.DownloadTask, error) {
	if strings.TrimSpace(taskID) == "" {
		return nil, errors.New("task_id empty")
	}

	task, err := u.downloadRepo.GetTaskByTaskID(taskID)
	if err != nil {
		return nil, errors.New("task not found")
	}
	if task.Status == "complete" {
		return task, errors.New("task already complete")
	}
	if task.Status == "stopped" {
		return task, nil
	}

	if task.Aria2GID != "" {
		if status, err := u.aria2Client.TellStatus(task.Aria2GID); err == nil {
			u.applyAria2Status(task, status)
			if fp := extractFilePath(status); fp != "" {
				task.FilePath = fp
			}
		}
		if _, err := u.aria2Client.Remove(task.Aria2GID); err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
			return task, err
		}
	}

	now := time.Now()
	task.Status = "stopped"
	task.ErrorMessage = "stopped by user"
	task.UpdatedAt = now
	if task.FinishedAt == nil {
		task.FinishedAt = &now
	}
	if err := u.downloadRepo.UpdateTask(task); err != nil {
		return nil, err
	}
	return task, nil
}

type StopTasksResult struct {
	Tasks  []*entity.DownloadTask `json:"tasks"`
	Failed map[string]string      `json:"failed"`
}

func (u *DownloadUseCase) StopTasks(taskIDs []string) StopTasksResult {
	result := StopTasksResult{
		Tasks:  []*entity.DownloadTask{},
		Failed: map[string]string{},
	}
	for _, taskID := range taskIDs {
		taskID = strings.TrimSpace(taskID)
		if taskID == "" {
			continue
		}
		task, err := u.StopTask(taskID)
		if err != nil {
			result.Failed[taskID] = err.Error()
			if task != nil {
				result.Tasks = append(result.Tasks, task)
			}
			continue
		}
		result.Tasks = append(result.Tasks, task)
	}
	return result
}

func (u *DownloadUseCase) DeleteTaskFile(taskID string) (*entity.DownloadTask, string, error) {
	if strings.TrimSpace(taskID) == "" {
		return nil, "", errors.New("task_id empty")
	}

	task, err := u.downloadRepo.GetTaskByTaskID(taskID)
	if err != nil {
		return nil, "", errors.New("task not found")
	}
	if isActiveDownloadStatus(task.Status) {
		return task, "", errors.New("stop task before deleting file")
	}

	filePath, err := u.getActualFilePath(task)
	if err != nil {
		return task, "", err
	}

	info, err := os.Stat(filePath)
	if err == nil {
		if info.IsDir() {
			return task, filePath, errors.New("refuse to delete directory")
		}
		if err := os.Remove(filePath); err != nil {
			return task, filePath, err
		}
		// Remove aria2 control file for stopped/partial downloads if it exists.
		_ = os.Remove(filePath + ".aria2")
	} else if !os.IsNotExist(err) {
		return task, filePath, err
	}

	now := time.Now()
	task.Status = "deleted"
	task.FilePath = ""
	task.ErrorMessage = ""
	task.Progress = 0
	task.UpdatedAt = now
	if err := u.downloadRepo.UpdateTask(task); err != nil {
		return nil, filePath, err
	}
	return task, filePath, nil
}

func (u *DownloadUseCase) RetryTask(deviceID string, taskID string) (*entity.DownloadTask, error) {
	task, err := u.downloadRepo.GetTaskByTaskID(taskID)
	if err != nil {
		return nil, errors.New("task not found")
	}
	if !isRetryableDownloadStatus(task.Status) {
		return nil, errors.New("only stopped or failed tasks can be retried")
	}

	directLink, err := u.storageUseCase.ResolveDirectLinkForDevice(deviceID, task.SourcePath)
	if err != nil {
		return nil, err
	}

	options := map[string]interface{}{}
	targetDir := strings.TrimSpace(u.config.Aria2.DownloadDir)
	if targetDir != "" {
		options["dir"] = targetDir
	}

	gid := ""
	if len(options) == 0 {
		gid, err = u.aria2Client.AddURI(directLink.DirectLink)
	} else {
		gid, err = u.aria2Client.AddURIWithOptions(directLink.DirectLink, options)
	}
	if err != nil {
		return nil, err
	}

	now := time.Now()
	task.DirectLink = directLink.DirectLink
	task.FileName = directLink.Name
	task.FileSize = directLink.Size
	task.Aria2GID = gid
	task.Status = "waiting"
	task.ErrorMessage = ""
	task.RetryCount++
	task.StartedAt = nil
	task.FinishedAt = nil
	task.UpdatedAt = now

	if err := u.downloadRepo.UpdateTask(task); err != nil {
		return nil, err
	}
	return task, nil
}

func isRetryableDownloadStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "error", "failed", "stopped", "cancelled", "canceled", "removed":
		return true
	default:
		return false
	}
}

func isActiveDownloadStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "waiting", "active", "paused", "submitted", "downloading":
		return true
	default:
		return false
	}
}

func (u *DownloadUseCase) applyAria2Status(task *entity.DownloadTask, status map[string]interface{}) {
	now := time.Now()
	if ariaStatus := stringFromAria2(status, "status"); ariaStatus != "" {
		if ariaStatus == "removed" {
			ariaStatus = "stopped"
		}
		task.Status = ariaStatus
	}

	task.CompletedLength = intFromAria2(status, "completedLength")
	task.TotalLength = intFromAria2(status, "totalLength")
	task.DownloadSpeed = intFromAria2(status, "downloadSpeed")
	if task.TotalLength <= 0 && task.FileSize > 0 {
		task.TotalLength = task.FileSize
	}

	if task.TotalLength > 0 && task.CompletedLength > 0 {
		task.Progress = math.Round(float64(task.CompletedLength)/float64(task.TotalLength)*10000) / 100
		if task.Status != "complete" && task.Progress >= 100 {
			task.Progress = 99.99
		}
	}
	if task.Status == "complete" {
		task.Progress = 100
		if task.CompletedLength <= 0 && task.TotalLength > 0 {
			task.CompletedLength = task.TotalLength
		}
		if task.FinishedAt == nil {
			task.FinishedAt = &now
		}
	}
	if (task.Status == "active" || task.Status == "waiting") && task.StartedAt == nil {
		task.StartedAt = &now
	}
	if task.Status == "error" {
		if message := stringFromAria2(status, "errorMessage"); message != "" {
			task.ErrorMessage = message
		}
	}
	if task.Status == "complete" && task.FilePath == "" {
		if fp := extractFilePath(status); fp != "" {
			task.FilePath = fp
		}
	}
	task.UpdatedAt = now
}

func stringFromAria2(status map[string]interface{}, key string) string {
	value, ok := status[key]
	if !ok || value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func intFromAria2(status map[string]interface{}, key string) int64 {
	value, ok := status[key]
	if !ok || value == nil {
		return 0
	}
	switch v := value.(type) {
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		return n
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	case json.Number:
		n, _ := v.Int64()
		return n
	default:
		return 0
	}
}

func (u *DownloadUseCase) getActualFilePath(task *entity.DownloadTask) (string, error) {
	// prefer persisted DB path (saved when aria2 completed)
	if task.FilePath != "" {
		normalized := filepath.FromSlash(task.FilePath)
		if filepath.IsAbs(normalized) {
			return normalized, nil
		}
	}

	// then try aria2 tellStatus for real path
	status, err := u.aria2Client.TellStatus(task.Aria2GID)
	if err == nil {
		ariaDir, _ := status["dir"].(string)

		if files, ok := status["files"].([]interface{}); ok && len(files) > 0 {
			if f, ok := files[0].(map[string]interface{}); ok {
				if path, ok := f["path"].(string); ok {
					path = strings.TrimSpace(path)
					if path != "" {
						normalized := filepath.FromSlash(path)
						if filepath.IsAbs(normalized) {
							return normalized, nil
						}
						if ariaDir != "" {
							return filepath.Join(filepath.FromSlash(ariaDir), normalized), nil
						}
						if u.config.Aria2.DownloadDir != "" {
							return filepath.Join(u.config.Aria2.DownloadDir, normalized), nil
						}
					}
				}
			}
		}
	}

	// fallback: config.DownloadDir + FileName
	if task.FileName != "" && u.config.Aria2.DownloadDir != "" {
		return filepath.Join(u.config.Aria2.DownloadDir, task.FileName), nil
	}

	return "", errors.New("cannot determine file path")
}

// extractFilePath extracts the file path from aria2 tellStatus response
func extractFilePath(status map[string]interface{}) string {
	ariaDir, _ := status["dir"].(string)
	if files, ok := status["files"].([]interface{}); ok && len(files) > 0 {
		if f, ok := files[0].(map[string]interface{}); ok {
			if path, ok := f["path"].(string); ok {
				path = strings.TrimSpace(path)
				if path == "" {
					return ""
				}
				normalized := filepath.FromSlash(path)
				if filepath.IsAbs(normalized) {
					return normalized
				}
				if ariaDir != "" {
					return filepath.Join(filepath.FromSlash(ariaDir), normalized)
				}
			}
		}
	}
	return ""
}

func (u *DownloadUseCase) OpenFile(taskID string) (string, error) {
	task, err := u.downloadRepo.GetTaskByTaskID(taskID)
	if err != nil {
		return "", errors.New("task not found")
	}
	if task.Status != "complete" {
		return "", errors.New("task not completed")
	}

	filePath, err := u.getActualFilePath(task)
	if err != nil {
		return "", err
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", filePath)
	case "darwin":
		cmd = exec.Command("open", filePath)
	default:
		cmd = exec.Command("xdg-open", filePath)
	}

	if err := cmd.Start(); err != nil {
		return "", errors.New("failed to open file: " + err.Error())
	}

	return filePath, nil
}

func (u *DownloadUseCase) OpenFileLocation(taskID string) (string, error) {
	task, err := u.downloadRepo.GetTaskByTaskID(taskID)
	if err != nil {
		return "", errors.New("task not found")
	}
	if task.Status != "complete" {
		return "", errors.New("task not completed")
	}

	filePath, err := u.getActualFilePath(task)
	if err != nil {
		return "", err
	}

	// verify the file exists; if not, search the download directory
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if found := findFileInDir(u.config.Aria2.DownloadDir, task.FileName); found != "" {
			filePath = found
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// Opening the parent folder is more reliable than /select,<file>
		// for names containing CJK punctuation or other shell-sensitive characters.
		cmd = exec.Command("explorer", filepath.Dir(filePath))
	case "darwin":
		cmd = exec.Command("open", "-R", filePath)
	default:
		cmd = exec.Command("xdg-open", filepath.Dir(filePath))
	}

	if err := cmd.Start(); err != nil {
		return "", errors.New("failed to open file location: " + err.Error())
	}

	return filepath.Dir(filePath), nil
}

// findFileInDir searches dir for a file matching filename (case-insensitive).
func findFileInDir(dir, filename string) string {
	if dir == "" || filename == "" {
		return ""
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() && strings.EqualFold(e.Name(), filename) {
			return filepath.Join(dir, e.Name())
		}
	}
	return ""
}

func (u *DownloadUseCase) CheckAria2Status() (map[string]interface{}, error) {
	version, err := u.aria2Client.GetVersion()
	if err != nil {
		return nil, err
	}
	stats, err := u.aria2Client.GetGlobalStat()
	if err == nil {
		if downloadSpeed, ok := stats["downloadSpeed"]; ok {
			version["downloadSpeed"] = downloadSpeed
		}
		if uploadSpeed, ok := stats["uploadSpeed"]; ok {
			version["uploadSpeed"] = uploadSpeed
		}
	}
	return version, nil
}

func newTaskID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405")))
	}
	return hex.EncodeToString(buf)
}
