package usecase

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"openbridge/backend/internal/config"
	"openbridge/backend/internal/domain/entity"
	"os"
	"time"

	"gorm.io/gorm"
)

const (
	userDataBackupFormat     = "openbridge-user-data-backup"
	userDataBackupVersion    = 1
	userDataBackupKDF        = "pbkdf2-sha256"
	userDataBackupIterations = 120000
)

var userDataBackupEnvKeys = []string{
	"APP_NAME",
	"APP_ENV",
	"APP_PORT",
	"APP_VERSION",
	"APP_AUTO_OPEN_BROWSER",
	"ARIA2_RPC_URL",
	"ARIA2_RPC_SECRET",
	"ARIA2_SECRET",
	"ARIA2_DOWNLOAD_DIR",
	"ARIA2_PATH",
	"ARIA2C_PATH",
	"ARIA2_AUTO_START",
	"OPENLIST_BASE_URL",
	"OPENLIST_TOKEN",
	"SESSION_DEVICE_LIMIT",
	"RCLONE_PATH",
	"FILETREE_CACHE_MAX_BYTES",
	"FILETREE_CACHE_DEPTH",
	"LOG_LEVEL",
	"LOG_FORMAT",
}

type UserDataBackupEnvelope struct {
	Format        string          `json:"format"`
	Version       int             `json:"version"`
	AppVersion    string          `json:"app_version"`
	CreatedAt     time.Time       `json:"created_at"`
	Encrypted     bool            `json:"encrypted"`
	KDF           string          `json:"kdf,omitempty"`
	KDFIterations int             `json:"kdf_iterations,omitempty"`
	Salt          string          `json:"salt,omitempty"`
	Nonce         string          `json:"nonce,omitempty"`
	Payload       json.RawMessage `json:"payload,omitempty"`
	Ciphertext    string          `json:"ciphertext,omitempty"`
}

type UserDataBackupPayload struct {
	ProviderAccounts []entity.ProviderAccount `json:"provider_accounts"`
	MountPoints      []entity.MountPoint      `json:"mount_points"`
	QuotaSnapshots   []entity.QuotaSnapshot   `json:"quota_snapshots"`
	DownloadTasks    []entity.DownloadTask    `json:"download_tasks"`
	RcloneProfiles   []entity.RcloneProfile   `json:"rclone_profiles"`
	Env              map[string]string        `json:"env"`
}

type UserDataRestoreResult struct {
	ProviderAccounts int  `json:"provider_accounts"`
	MountPoints      int  `json:"mount_points"`
	QuotaSnapshots   int  `json:"quota_snapshots"`
	DownloadTasks    int  `json:"download_tasks"`
	RcloneProfiles   int  `json:"rclone_profiles"`
	EnvSettings      int  `json:"env_settings"`
	RestartRequired  bool `json:"restart_required"`
}

func (uc *UserUseCase) Backup(password string) ([]byte, string, error) {
	payload, err := uc.buildBackupPayload()
	if err != nil {
		return nil, "", err
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, "", err
	}

	envelope := UserDataBackupEnvelope{
		Format:     userDataBackupFormat,
		Version:    userDataBackupVersion,
		AppVersion: normalizeAppVersion(uc.config.App.Version),
		CreatedAt:  time.Now(),
		Encrypted:  password != "",
	}

	if password == "" {
		envelope.Payload = payloadBytes
	} else {
		ciphertext, salt, nonce, err := encryptBackupPayload(payloadBytes, password)
		if err != nil {
			return nil, "", err
		}
		envelope.KDF = userDataBackupKDF
		envelope.KDFIterations = userDataBackupIterations
		envelope.Salt = base64.StdEncoding.EncodeToString(salt)
		envelope.Nonce = base64.StdEncoding.EncodeToString(nonce)
		envelope.Ciphertext = base64.StdEncoding.EncodeToString(ciphertext)
	}

	backupBytes, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("openbridge-user-data-backup-%s.json", time.Now().Format("20060102-150405"))
	return backupBytes, filename, nil
}

func (uc *UserUseCase) Restore(backupBytes []byte, password string) (UserDataRestoreResult, error) {
	var envelope UserDataBackupEnvelope
	if err := json.Unmarshal(backupBytes, &envelope); err != nil {
		return UserDataRestoreResult{}, err
	}
	if envelope.Format != userDataBackupFormat {
		return UserDataRestoreResult{}, errors.New("unsupported backup format")
	}
	if envelope.Version != userDataBackupVersion {
		return UserDataRestoreResult{}, fmt.Errorf("unsupported backup version %d", envelope.Version)
	}

	payloadBytes, err := decodeBackupPayload(envelope, password)
	if err != nil {
		return UserDataRestoreResult{}, err
	}

	var payload UserDataBackupPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return UserDataRestoreResult{}, err
	}

	result := UserDataRestoreResult{
		ProviderAccounts: len(payload.ProviderAccounts),
		MountPoints:      len(payload.MountPoints),
		QuotaSnapshots:   len(payload.QuotaSnapshots),
		DownloadTasks:    len(payload.DownloadTasks),
		RcloneProfiles:   len(payload.RcloneProfiles),
		RestartRequired:  true,
	}

	if err := uc.restoreBackupTables(payload); err != nil {
		return UserDataRestoreResult{}, err
	}

	envCount, err := uc.restoreBackupEnv(payload.Env)
	if err != nil {
		return UserDataRestoreResult{}, err
	}
	result.EnvSettings = envCount
	uc.clearSessions()

	return result, nil
}

func (uc *UserUseCase) buildBackupPayload() (UserDataBackupPayload, error) {
	payload := UserDataBackupPayload{}

	if err := uc.db.Find(&payload.ProviderAccounts).Error; err != nil {
		return payload, err
	}
	if err := uc.db.Find(&payload.MountPoints).Error; err != nil {
		return payload, err
	}
	if err := uc.db.Find(&payload.QuotaSnapshots).Error; err != nil {
		return payload, err
	}
	if err := uc.db.Find(&payload.DownloadTasks).Error; err != nil {
		return payload, err
	}
	if err := uc.db.Find(&payload.RcloneProfiles).Error; err != nil {
		return payload, err
	}

	env, err := collectBackupEnv()
	if err != nil {
		return payload, err
	}
	payload.Env = env

	return payload, nil
}

func (uc *UserUseCase) restoreBackupTables(payload UserDataBackupPayload) error {
	return uc.db.Transaction(func(tx *gorm.DB) error {
		for _, table := range []string{
			"quota_snapshots",
			"rclone_profiles",
			"mount_points",
			"provider_accounts",
			"download_tasks",
		} {
			if err := tx.Exec("DELETE FROM " + table).Error; err != nil {
				return err
			}
		}

		if len(payload.ProviderAccounts) > 0 {
			if err := tx.Create(&payload.ProviderAccounts).Error; err != nil {
				return err
			}
		}
		if len(payload.MountPoints) > 0 {
			if err := tx.Create(&payload.MountPoints).Error; err != nil {
				return err
			}
		}
		if len(payload.QuotaSnapshots) > 0 {
			if err := tx.Create(&payload.QuotaSnapshots).Error; err != nil {
				return err
			}
		}
		if len(payload.DownloadTasks) > 0 {
			if err := tx.Create(&payload.DownloadTasks).Error; err != nil {
				return err
			}
		}
		if len(payload.RcloneProfiles) > 0 {
			if err := tx.Create(&payload.RcloneProfiles).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func collectBackupEnv() (map[string]string, error) {
	values, err := config.ReadEnvFile()
	if err != nil {
		return nil, err
	}

	snapshot := make(map[string]string, len(userDataBackupEnvKeys))
	for _, key := range userDataBackupEnvKeys {
		if value, ok := values[key]; ok {
			snapshot[key] = value
			continue
		}
		if value := os.Getenv(key); value != "" {
			snapshot[key] = value
		}
	}

	return snapshot, nil
}

func (uc *UserUseCase) restoreBackupEnv(values map[string]string) (int, error) {
	if len(values) == 0 {
		return 0, nil
	}

	updates := make(map[string]string, len(userDataBackupEnvKeys))
	for _, key := range userDataBackupEnvKeys {
		if value, ok := values[key]; ok {
			updates[key] = value
		}
	}
	if len(updates) == 0 {
		return 0, nil
	}

	if err := config.SetEnvValues(updates); err != nil {
		return 0, err
	}

	refreshed := config.ReadConfig()
	*uc.config = refreshed

	return len(updates), nil
}

func decodeBackupPayload(envelope UserDataBackupEnvelope, password string) ([]byte, error) {
	if !envelope.Encrypted {
		if len(envelope.Payload) == 0 {
			return nil, errors.New("backup payload is empty")
		}
		return envelope.Payload, nil
	}

	if password == "" {
		return nil, errors.New("backup password is required")
	}
	if envelope.KDF != userDataBackupKDF {
		return nil, errors.New("unsupported backup key derivation")
	}
	if envelope.KDFIterations <= 0 {
		return nil, errors.New("invalid backup key derivation iterations")
	}
	if envelope.KDFIterations > 1000000 {
		return nil, errors.New("backup key derivation iterations too large")
	}

	salt, err := base64.StdEncoding.DecodeString(envelope.Salt)
	if err != nil {
		return nil, err
	}
	if len(salt) == 0 {
		return nil, errors.New("backup salt is empty")
	}
	nonce, err := base64.StdEncoding.DecodeString(envelope.Nonce)
	if err != nil {
		return nil, err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(envelope.Ciphertext)
	if err != nil {
		return nil, err
	}

	key := pbkdf2SHA256([]byte(password), salt, envelope.KDFIterations, 32)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(nonce) != aead.NonceSize() {
		return nil, errors.New("invalid backup nonce size")
	}

	payload, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("backup password is incorrect or file is corrupted")
	}

	return payload, nil
}

func encryptBackupPayload(payload []byte, password string) ([]byte, []byte, []byte, error) {
	salt, err := randomBytes(16)
	if err != nil {
		return nil, nil, nil, err
	}
	nonce, err := randomBytes(12)
	if err != nil {
		return nil, nil, nil, err
	}

	key := pbkdf2SHA256([]byte(password), salt, userDataBackupIterations, 32)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, err
	}

	return aead.Seal(nil, nonce, payload, nil), salt, nonce, nil
}

func randomBytes(length int) ([]byte, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func pbkdf2SHA256(password, salt []byte, iterations, keyLen int) []byte {
	hashLen := sha256.Size
	blockCount := (keyLen + hashLen - 1) / hashLen
	key := make([]byte, 0, blockCount*hashLen)

	for blockIndex := 1; blockIndex <= blockCount; blockIndex++ {
		mac := hmac.New(sha256.New, password)
		_, _ = mac.Write(salt)

		var intBlock [4]byte
		binary.BigEndian.PutUint32(intBlock[:], uint32(blockIndex))
		_, _ = mac.Write(intBlock[:])

		sum := mac.Sum(nil)
		block := make([]byte, len(sum))
		copy(block, sum)

		for i := 1; i < iterations; i++ {
			mac = hmac.New(sha256.New, password)
			_, _ = mac.Write(sum)
			sum = mac.Sum(nil)
			for j := range block {
				block[j] ^= sum[j]
			}
		}

		key = append(key, block...)
	}

	return key[:keyLen]
}

func (uc *UserUseCase) clearSessions() {
	uc.sessionMu.Lock()
	defer uc.sessionMu.Unlock()
	uc.sessions = make(map[string]*deviceSession)
}
