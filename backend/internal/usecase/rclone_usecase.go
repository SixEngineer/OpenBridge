package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"openbridge/backend/internal/config"
	"openbridge/backend/internal/domain/entity"
	"openbridge/backend/internal/repository"
	"openbridge/backend/internal/tool"
)

var rcloneNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)

type RcloneUseCase struct {
	config       *config.Config
	repo         *repository.RcloneProfileRepository
	mountRepo    *repository.MountRepository
	providerRepo *repository.ProviderRepository
	secretCodec  *tool.SecretCodec
}

type RcloneProfileInput struct {
	Name       string `json:"name"`
	Mode       string `json:"mode"`
	MountIDs   []uint `json:"mount_ids"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	TargetPath string `json:"target_path"`
}

type RcloneProfileView struct {
	ID            uint       `json:"id"`
	Name          string     `json:"name"`
	Mode          string     `json:"mode"`
	MountIDs      []uint     `json:"mount_ids"`
	Username      string     `json:"username"`
	TargetPath    string     `json:"target_path"`
	PasswordSaved bool       `json:"password_saved"`
	LastAppliedAt *time.Time `json:"last_applied_at"`
	LastMountedAt *time.Time `json:"last_mounted_at"`
	LastError     string     `json:"last_error"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ApplyCommands []string   `json:"apply_commands"`
	MountCommand  string     `json:"mount_command"`
}

func NewRcloneUseCase(
	cfg *config.Config,
	repo *repository.RcloneProfileRepository,
	mountRepo *repository.MountRepository,
	providerRepo *repository.ProviderRepository,
) *RcloneUseCase {
	seed := strings.Join([]string{cfg.App.Name, cfg.DB.Path, cfg.App.Port}, "|")
	return &RcloneUseCase{
		config:       cfg,
		repo:         repo,
		mountRepo:    mountRepo,
		providerRepo: providerRepo,
		secretCodec:  tool.NewSecretCodec(seed),
	}
}

func (u *RcloneUseCase) ListProfiles() ([]RcloneProfileView, error) {
	profiles, err := u.repo.List()
	if err != nil {
		return nil, err
	}

	result := make([]RcloneProfileView, 0, len(profiles))
	for _, profile := range profiles {
		view, err := u.toView(&profile)
		if err != nil {
			return nil, err
		}
		result = append(result, view)
	}
	return result, nil
}

func (u *RcloneUseCase) CreateProfile(input RcloneProfileInput) (RcloneProfileView, error) {
	if err := u.validateInput(input, true); err != nil {
		return RcloneProfileView{}, err
	}

	mountIDsJSON, err := json.Marshal(uniqueUintSlice(input.MountIDs))
	if err != nil {
		return RcloneProfileView{}, err
	}

	passwordCipher, err := u.secretCodec.Encrypt(strings.TrimSpace(input.Password))
	if err != nil {
		return RcloneProfileView{}, err
	}

	profile := &entity.RcloneProfile{
		Name:           strings.TrimSpace(input.Name),
		Mode:           strings.TrimSpace(input.Mode),
		MountIDs:       string(mountIDsJSON),
		Username:       strings.TrimSpace(input.Username),
		PasswordCipher: passwordCipher,
		TargetPath:     strings.TrimSpace(input.TargetPath),
	}

	if err := u.repo.Create(profile); err != nil {
		return RcloneProfileView{}, err
	}

	return u.toView(profile)
}

func (u *RcloneUseCase) UpdateProfile(id uint, input RcloneProfileInput) (RcloneProfileView, error) {
	profile, err := u.repo.Get(id)
	if err != nil {
		return RcloneProfileView{}, err
	}

	if err := u.validateInput(input, false); err != nil {
		return RcloneProfileView{}, err
	}

	if strings.TrimSpace(input.Name) != "" {
		profile.Name = strings.TrimSpace(input.Name)
	}
	if strings.TrimSpace(input.Mode) != "" {
		profile.Mode = strings.TrimSpace(input.Mode)
	}
	if len(input.MountIDs) > 0 {
		mountIDsJSON, err := json.Marshal(uniqueUintSlice(input.MountIDs))
		if err != nil {
			return RcloneProfileView{}, err
		}
		profile.MountIDs = string(mountIDsJSON)
	}
	if strings.TrimSpace(input.Username) != "" {
		profile.Username = strings.TrimSpace(input.Username)
	}
	if strings.TrimSpace(input.TargetPath) != "" {
		profile.TargetPath = strings.TrimSpace(input.TargetPath)
	}
	if strings.TrimSpace(input.Password) != "" {
		passwordCipher, err := u.secretCodec.Encrypt(strings.TrimSpace(input.Password))
		if err != nil {
			return RcloneProfileView{}, err
		}
		profile.PasswordCipher = passwordCipher
	}

	if err := u.repo.Update(profile); err != nil {
		return RcloneProfileView{}, err
	}

	return u.toView(profile)
}

func (u *RcloneUseCase) DeleteProfile(id uint) error {
	return u.repo.Delete(id)
}

func (u *RcloneUseCase) ApplyProfile(id uint) (RcloneProfileView, error) {
	profile, err := u.repo.Get(id)
	if err != nil {
		return RcloneProfileView{}, err
	}

	password, err := u.secretCodec.Decrypt(profile.PasswordCipher)
	if err != nil {
		return RcloneProfileView{}, err
	}

	mounts, err := u.resolveMounts(profile)
	if err != nil {
		return RcloneProfileView{}, err
	}

	if err := u.ensureRcloneAvailable(); err != nil {
		return RcloneProfileView{}, err
	}

	if err := u.applyProfileConfig(profile, mounts, password); err != nil {
		profile.LastError = err.Error()
		_ = u.repo.Update(profile)
		return RcloneProfileView{}, err
	}

	now := time.Now()
	profile.LastAppliedAt = &now
	profile.LastError = ""
	if err := u.repo.Update(profile); err != nil {
		return RcloneProfileView{}, err
	}

	return u.toView(profile)
}

func (u *RcloneUseCase) StartMount(id uint) (RcloneProfileView, error) {
	profile, err := u.repo.Get(id)
	if err != nil {
		return RcloneProfileView{}, err
	}

	if _, err := u.ApplyProfile(id); err != nil {
		return RcloneProfileView{}, err
	}

	if err := u.startMountProcess(profile); err != nil {
		profile.LastError = err.Error()
		_ = u.repo.Update(profile)
		return RcloneProfileView{}, err
	}

	now := time.Now()
	profile.LastMountedAt = &now
	profile.LastError = ""
	if err := u.repo.Update(profile); err != nil {
		return RcloneProfileView{}, err
	}

	return u.toView(profile)
}

func (u *RcloneUseCase) toView(profile *entity.RcloneProfile) (RcloneProfileView, error) {
	mountIDs, err := parseMountIDs(profile.MountIDs)
	if err != nil {
		return RcloneProfileView{}, err
	}

	mounts, err := u.resolveMounts(profile)
	if err != nil {
		mounts = nil
	}

	return RcloneProfileView{
		ID:            profile.ID,
		Name:          profile.Name,
		Mode:          profile.Mode,
		MountIDs:      mountIDs,
		Username:      profile.Username,
		TargetPath:    profile.TargetPath,
		PasswordSaved: strings.TrimSpace(profile.PasswordCipher) != "",
		LastAppliedAt: profile.LastAppliedAt,
		LastMountedAt: profile.LastMountedAt,
		LastError:     profile.LastError,
		CreatedAt:     profile.CreatedAt,
		UpdatedAt:     profile.UpdatedAt,
		ApplyCommands: u.previewApplyCommands(profile, mounts),
		MountCommand:  u.previewMountCommand(profile),
	}, nil
}

func (u *RcloneUseCase) validateInput(input RcloneProfileInput, requirePassword bool) error {
	name := strings.TrimSpace(input.Name)
	mode := strings.TrimSpace(input.Mode)
	username := strings.TrimSpace(input.Username)
	targetPath := strings.TrimSpace(input.TargetPath)

	if name == "" {
		return errors.New("rclone profile name required")
	}
	if !rcloneNamePattern.MatchString(name) {
		return errors.New("rclone profile name must start with a letter or number and use only letters, numbers, underscore, or hyphen")
	}
	if mode != "ordinary" && mode != "union" && mode != "combine" {
		return errors.New("rclone mode invalid")
	}
	if len(input.MountIDs) == 0 {
		return errors.New("at least one mount id required")
	}
	if mode == "ordinary" && len(input.MountIDs) != 1 {
		return errors.New("ordinary mode requires exactly one mount id")
	}
	if (mode == "union" || mode == "combine") && len(input.MountIDs) < 2 {
		return errors.New("union/combine mode requires at least two mount ids")
	}
	if username == "" {
		return errors.New("rclone username required")
	}
	if requirePassword && strings.TrimSpace(input.Password) == "" {
		return errors.New("rclone password required")
	}
	if targetPath == "" {
		return errors.New("rclone target path required")
	}
	return nil
}

func (u *RcloneUseCase) resolveMounts(profile *entity.RcloneProfile) ([]entity.MountPoint, error) {
	mountIDs, err := parseMountIDs(profile.MountIDs)
	if err != nil {
		return nil, err
	}

	scope := config.NormalizeBaseURLScope(u.config.OpenList.BaseURL)
	mounts := make([]entity.MountPoint, 0, len(mountIDs))
	for _, mountID := range mountIDs {
		mount, err := u.mountRepo.GetMountPoint(mountID)
		if err != nil {
			return nil, fmt.Errorf("mount %d not found", mountID)
		}
		if !mount.Enabled {
			return nil, fmt.Errorf("mount %d is disabled", mountID)
		}
		if _, err := u.providerRepo.GetProviderAccountByOpenList(mount.ProviderAccountID, scope); err != nil {
			return nil, fmt.Errorf("mount %d does not belong to the current openlist source", mountID)
		}
		mounts = append(mounts, *mount)
	}
	return mounts, nil
}

func (u *RcloneUseCase) ensureRcloneAvailable() error {
	_, err := u.runRcloneCommand(10*time.Second, "version")
	return err
}

func (u *RcloneUseCase) applyProfileConfig(profile *entity.RcloneProfile, mounts []entity.MountPoint, password string) error {
	if profile.Mode == "ordinary" {
		mount := mounts[0]
		return u.createWebDAVRemote(profile.Name, mount.ID, profile.Username, password)
	}

	for _, mount := range mounts {
		underlyingRemote := u.underlyingRemoteName(profile.Name, mount.ID)
		if err := u.createWebDAVRemote(underlyingRemote, mount.ID, profile.Username, password); err != nil {
			return err
		}
	}

	switch profile.Mode {
	case "union":
		return u.createUnionRemote(profile.Name, mounts)
	case "combine":
		return u.createCombineRemote(profile.Name, mounts)
	default:
		return errors.New("unsupported rclone mode")
	}
}

func (u *RcloneUseCase) createWebDAVRemote(remoteName string, mountID uint, username string, password string) error {
	_ = u.deleteRemote(remoteName)
	args := []string{
		"config", "create", remoteName, "webdav",
		"url", u.mountWebDAVURL(mountID),
		"vendor", "other",
		"user", username,
		"pass", password,
		"--obscure",
	}
	_, err := u.runRcloneCommand(20*time.Second, args...)
	return err
}

func (u *RcloneUseCase) createUnionRemote(remoteName string, mounts []entity.MountPoint) error {
	_ = u.deleteRemote(remoteName)

	upstreams := make([]string, 0, len(mounts))
	for _, mount := range mounts {
		upstreams = append(upstreams, u.underlyingRemoteName(remoteName, mount.ID)+":")
	}

	args := []string{
		"config", "create", remoteName, "union",
		"upstreams", strings.Join(upstreams, " "),
		"action_policy", "epall",
		"create_policy", "epmfs",
		"search_policy", "ff",
	}
	_, err := u.runRcloneCommand(20*time.Second, args...)
	return err
}

func (u *RcloneUseCase) createCombineRemote(remoteName string, mounts []entity.MountPoint) error {
	_ = u.deleteRemote(remoteName)

	upstreams := make([]string, 0, len(mounts))
	usedNames := map[string]int{}
	for _, mount := range mounts {
		dirName := sanitizeCombineDirName(mount.Name)
		if dirName == "" {
			dirName = fmt.Sprintf("mount_%d", mount.ID)
		}
		if count := usedNames[dirName]; count > 0 {
			dirName = fmt.Sprintf("%s_%d", dirName, mount.ID)
		}
		usedNames[dirName]++
		upstreams = append(upstreams, dirName+"="+u.underlyingRemoteName(remoteName, mount.ID)+":")
	}

	args := []string{
		"config", "create", remoteName, "combine",
		"upstreams", strings.Join(upstreams, " "),
	}
	_, err := u.runRcloneCommand(20*time.Second, args...)
	return err
}

func (u *RcloneUseCase) deleteRemote(remoteName string) error {
	_, err := u.runRcloneCommand(10*time.Second, "config", "delete", remoteName)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "didn't find") && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		return err
	}
	return nil
}

func (u *RcloneUseCase) previewApplyCommands(profile *entity.RcloneProfile, mounts []entity.MountPoint) []string {
	rcloneBin := quoteCLI(displayRclonePath(u.config.Rclone.Path))
	if len(mounts) == 0 {
		return []string{}
	}

	commands := []string{}
	if profile.Mode == "ordinary" {
		commands = append(commands, fmt.Sprintf(
			`%s config create %s webdav url %s vendor other user %s pass ******** --obscure`,
			rcloneBin,
			profile.Name,
			quoteCLI(u.mountWebDAVURL(mounts[0].ID)),
			quoteCLI(profile.Username),
		))
		return commands
	}

	for _, mount := range mounts {
		commands = append(commands, fmt.Sprintf(
			`%s config create %s webdav url %s vendor other user %s pass ******** --obscure`,
			rcloneBin,
			u.underlyingRemoteName(profile.Name, mount.ID),
			quoteCLI(u.mountWebDAVURL(mount.ID)),
			quoteCLI(profile.Username),
		))
	}

	switch profile.Mode {
	case "union":
		upstreams := make([]string, 0, len(mounts))
		for _, mount := range mounts {
			upstreams = append(upstreams, u.underlyingRemoteName(profile.Name, mount.ID)+":")
		}
		commands = append(commands, fmt.Sprintf(
			`%s config create %s union upstreams %s action_policy epall create_policy epmfs search_policy ff`,
			rcloneBin,
			profile.Name,
			quoteCLI(strings.Join(upstreams, " ")),
		))
	case "combine":
		upstreams := make([]string, 0, len(mounts))
		usedNames := map[string]int{}
		for _, mount := range mounts {
			dirName := sanitizeCombineDirName(mount.Name)
			if dirName == "" {
				dirName = fmt.Sprintf("mount_%d", mount.ID)
			}
			if count := usedNames[dirName]; count > 0 {
				dirName = fmt.Sprintf("%s_%d", dirName, mount.ID)
			}
			usedNames[dirName]++
			upstreams = append(upstreams, dirName+"="+u.underlyingRemoteName(profile.Name, mount.ID)+":")
		}
		commands = append(commands, fmt.Sprintf(
			`%s config create %s combine upstreams %s`,
			rcloneBin,
			profile.Name,
			quoteCLI(strings.Join(upstreams, " ")),
		))
	}
	return commands
}

func (u *RcloneUseCase) previewMountCommand(profile *entity.RcloneProfile) string {
	rcloneBin := quoteCLI(displayRclonePath(u.config.Rclone.Path))
	return fmt.Sprintf(
		`%s mount %s: %s --vfs-cache-mode full%s`,
		rcloneBin,
		profile.Name,
		quoteCLI(profile.TargetPath),
		defaultMountFlagsPreview(),
	)
}

func (u *RcloneUseCase) startMountProcess(profile *entity.RcloneProfile) error {
	rclonePath := resolvedRclonePath(u.config.Rclone.Path)
	args := []string{"mount", profile.Name + ":", profile.TargetPath, "--vfs-cache-mode", "full"}
	if runtime.GOOS == "windows" {
		args = append(args, "--links")
		cmdArgs := append([]string{"/c", "start", "", "/b", rclonePath}, args...)
		cmd := exec.Command("cmd", cmdArgs...)
		return cmd.Start()
	}

	if runtime.GOOS != "windows" {
		args = append(args, "--daemon")
	}
	cmd := exec.Command(rclonePath, args...)
	return cmd.Start()
}

func (u *RcloneUseCase) runRcloneCommand(timeout time.Duration, args ...string) (string, error) {
	rclonePath := resolvedRclonePath(u.config.Rclone.Path)
	cmd := exec.Command(rclonePath, args...)

	done := make(chan struct{})
	var output []byte
	var err error
	go func() {
		output, err = cmd.CombinedOutput()
		close(done)
	}()

	select {
	case <-done:
		if err != nil {
			return "", fmt.Errorf("%s", strings.TrimSpace(string(output)))
		}
		return strings.TrimSpace(string(output)), nil
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		return "", errors.New("rclone command timed out")
	}
}

func (u *RcloneUseCase) mountWebDAVURL(mountID uint) string {
	return fmt.Sprintf("http://127.0.0.1:%s/api/v1/webdav/mounts/%d", u.config.App.Port, mountID)
}

func (u *RcloneUseCase) underlyingRemoteName(profileName string, mountID uint) string {
	return fmt.Sprintf("%s__mount_%d", profileName, mountID)
}

func parseMountIDs(raw string) ([]uint, error) {
	var ids []uint
	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func uniqueUintSlice(values []uint) []uint {
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if value == 0 || containsUint(result, value) {
			continue
		}
		result = append(result, value)
	}
	return result
}

func containsUint(values []uint, target uint) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func sanitizeCombineDirName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return strings.Trim(b.String(), "_")
}

func resolvedRclonePath(value string) string {
	if strings.TrimSpace(value) == "" {
		return "rclone"
	}
	return strings.TrimSpace(value)
}

func displayRclonePath(value string) string {
	if strings.TrimSpace(value) == "" {
		return "rclone"
	}
	return strings.TrimSpace(value)
}

func quoteCLI(value string) string {
	if value == "" {
		return `""`
	}
	if strings.ContainsAny(value, " \t") || strings.Contains(value, `\`) || strings.Contains(value, ":") || strings.Contains(value, "/") {
		return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
	}
	return value
}

func defaultMountFlagsPreview() string {
	if runtime.GOOS == "windows" {
		return " --links"
	}
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		return " --daemon"
	}
	return ""
}
