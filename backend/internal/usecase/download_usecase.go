package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

func (u *DownloadUseCase) CreateTask(path string, dir string) (*entity.DownloadTask, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("path empty")
	}

	directLink, err := u.storageUseCase.ResolveDirectLink(path)
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
		SourcePath:      path,
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

	if status, err := u.aria2Client.TellStatus(task.Aria2GID); err == nil {
		task.Status = status["status"].(string)
		if task.Status == "complete" && task.FinishedAt == nil {
			now := time.Now()
			task.FinishedAt = &now
		}
		// persist actual file path when task completes, so aria2 forgetting it later doesn't matter
		if task.Status == "complete" && task.FilePath == "" {
			if fp := extractFilePath(status); fp != "" {
				task.FilePath = fp
			}
		}
		u.downloadRepo.UpdateTask(task)
	} else if strings.Contains(err.Error(), "not found") && task.Status != "complete" {
		// GID not found in aria2 -> task was removed or expired -> mark as error.
		task.Status = "error"
		u.downloadRepo.UpdateTask(task)
	} // Network errors (aria2 unreachable) or already complete: keep current DB status.

	return task, nil
}

func (u *DownloadUseCase) RetryTask(taskID string) (*entity.DownloadTask, error) {
	task, err := u.downloadRepo.GetTaskByTaskID(taskID)
	if err != nil {
		return nil, errors.New("task not found")
	}

	directLink, err := u.storageUseCase.ResolveDirectLink(task.SourcePath)
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
	return version, nil
}

func newTaskID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405")))
	}
	return hex.EncodeToString(buf)
}
