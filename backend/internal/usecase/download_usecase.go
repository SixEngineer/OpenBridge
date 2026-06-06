package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
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
		TaskID:     newTaskID(),
		SourcePath: path,
		DirectLink: directLink.DirectLink,
		FileName:   directLink.Name,
		FileSize:   directLink.Size,
		Aria2GID:   gid,
		Status:     "waiting",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
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
		u.downloadRepo.UpdateTask(task)
	} else if strings.Contains(err.Error(), "not found") && task.Status != "complete" {
		// GID not found in aria2 → task was removed or expired → mark as error.
		task.Status = "error"
		u.downloadRepo.UpdateTask(task)
	} // Network errors (aria2 unreachable) or already complete: keep current DB status.

	return task, nil
}

func newTaskID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405")))
	}
	return hex.EncodeToString(buf)
}
