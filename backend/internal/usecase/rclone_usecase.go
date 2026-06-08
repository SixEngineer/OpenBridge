package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
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
	legacyCodecs []*tool.SecretCodec
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
	IsMounted     bool       `json:"is_mounted"`
	MountPID      int        `json:"mount_pid"`
	MountRCAddr   string     `json:"mount_rc_addr"`
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
	seed := strings.Join([]string{cfg.App.Name, cfg.DB.Path}, "|")
	legacySeed := strings.Join([]string{cfg.App.Name, cfg.DB.Path, cfg.App.Port}, "|")
	legacyCodecs := []*tool.SecretCodec{}
	if legacySeed != seed {
		legacyCodecs = append(legacyCodecs, tool.NewSecretCodec(legacySeed))
	}
	return &RcloneUseCase{
		config:       cfg,
		repo:         repo,
		mountRepo:    mountRepo,
		providerRepo: providerRepo,
		secretCodec:  tool.NewSecretCodec(seed),
		legacyCodecs: legacyCodecs,
	}
}

func (u *RcloneUseCase) ListProfiles() ([]RcloneProfileView, error) {
	if err := u.repo.AssignEmptyOpenListScope(u.currentOpenListScope()); err != nil {
		return nil, err
	}
	_ = u.syncProfiles()
	profiles, err := u.repo.List(u.currentOpenListScope())
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
	if err := u.repo.AssignEmptyOpenListScope(u.currentOpenListScope()); err != nil {
		return RcloneProfileView{}, err
	}
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
		OpenListBaseURL: u.currentOpenListScope(),
		Name:            strings.TrimSpace(input.Name),
		Mode:            strings.TrimSpace(input.Mode),
		MountIDs:        string(mountIDsJSON),
		ManagedRemotes:  "[]",
		Username:        strings.TrimSpace(input.Username),
		PasswordCipher:  passwordCipher,
		TargetPath:      strings.TrimSpace(input.TargetPath),
	}

	if err := u.repo.Create(profile); err != nil {
		return RcloneProfileView{}, err
	}

	return u.toView(profile)
}

func (u *RcloneUseCase) UpdateProfile(id uint, input RcloneProfileInput) (RcloneProfileView, error) {
	if err := u.repo.AssignEmptyOpenListScope(u.currentOpenListScope()); err != nil {
		return RcloneProfileView{}, err
	}
	profile, err := u.repo.Get(id, u.currentOpenListScope())
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
	if err := u.repo.AssignEmptyOpenListScope(u.currentOpenListScope()); err != nil {
		return err
	}
	profile, err := u.repo.Get(id, u.currentOpenListScope())
	if err != nil {
		return err
	}
	if err := u.stopMountedProfile(profile); err != nil {
		return err
	}
	managedRemotes, _ := parseStringSlice(profile.ManagedRemotes)
	if len(managedRemotes) == 0 {
		mountIDs, _ := parseMountIDs(profile.MountIDs)
		managedRemotes = u.legacyRemoteNamesFromIDs(profile.Name, profile.Mode, mountIDs)
	}
	if err := u.cleanupManagedRemotes(managedRemotes, nil); err != nil {
		return err
	}
	return u.repo.Delete(id, u.currentOpenListScope())
}

func (u *RcloneUseCase) ApplyProfile(id uint) (RcloneProfileView, error) {
	profile, err := u.repo.Get(id, u.currentOpenListScope())
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

	password, migratedCipher, err := u.decryptProfilePassword(profile, mounts, profile.PasswordCipher)
	if err == nil && migratedCipher {
		passwordCipher, encryptErr := u.secretCodec.Encrypt(password)
		if encryptErr != nil {
			return RcloneProfileView{}, encryptErr
		}
		profile.PasswordCipher = passwordCipher
	}
	if err == nil {
		err = u.applyProfileConfig(profile, mounts, password)
	} else {
		err = u.applyProfileConfigWithoutPassword(profile, mounts)
		if err != nil {
			err = fmt.Errorf("%s; %w", u.passwordRecoveryHint(), err)
		}
	}
	if err != nil {
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
	profile, err := u.repo.Get(id, u.currentOpenListScope())
	if err != nil {
		return RcloneProfileView{}, err
	}
	if profile.IsMounted {
		if err := u.stopMountedProfile(profile); err != nil {
			return RcloneProfileView{}, err
		}
	}

	if _, err := u.ApplyProfile(id); err != nil {
		return RcloneProfileView{}, err
	}

	pid, rcAddr, err := u.startMountProcess(profile)
	if err != nil {
		profile.LastError = err.Error()
		_ = u.repo.Update(profile)
		return RcloneProfileView{}, err
	}

	now := time.Now()
	profile.IsMounted = true
	profile.MountPID = pid
	profile.MountRCAddr = rcAddr
	profile.LastMountedAt = &now
	profile.LastError = ""
	if err := u.repo.Update(profile); err != nil {
		return RcloneProfileView{}, err
	}

	return u.toView(profile)
}

func (u *RcloneUseCase) StopMount(id uint) (RcloneProfileView, error) {
	profile, err := u.repo.Get(id, u.currentOpenListScope())
	if err != nil {
		return RcloneProfileView{}, err
	}
	if err := u.stopMountedProfile(profile); err != nil {
		profile.LastError = err.Error()
		_ = u.repo.Update(profile)
		return RcloneProfileView{}, err
	}
	if err := u.repo.Update(profile); err != nil {
		return RcloneProfileView{}, err
	}
	return u.toView(profile)
}

func (u *RcloneUseCase) HandleMountDeleted(mountID uint) error {
	profiles, err := u.repo.List(u.currentOpenListScope())
	if err != nil {
		return err
	}
	for i := range profiles {
		profile := profiles[i]
		mountIDs, err := parseMountIDs(profile.MountIDs)
		if err != nil {
			continue
		}
		if !containsUint(mountIDs, mountID) {
			continue
		}
		if err := u.reconcileProfileAfterMountChange(&profile, mountID); err != nil {
			return err
		}
	}
	return nil
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
		IsMounted:     profile.IsMounted,
		MountPID:      profile.MountPID,
		MountRCAddr:   profile.MountRCAddr,
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
	managedRemotes, _ := parseStringSlice(profile.ManagedRemotes)
	desiredRemotes := u.expectedRemoteNames(profile, mounts)
	if err := u.cleanupManagedRemotes(managedRemotes, desiredRemotes); err != nil {
		return err
	}

	if profile.Mode == "ordinary" {
		mount := mounts[0]
		if err := u.createWebDAVRemote(profile.Name, mount.ID, profile.Username, password); err != nil {
			return err
		}
		return u.updateManagedRemotes(profile, desiredRemotes)
	}

	for _, mount := range mounts {
		underlyingRemote := u.underlyingRemoteName(profile.Name, mount.ID)
		if err := u.createWebDAVRemote(underlyingRemote, mount.ID, profile.Username, password); err != nil {
			return err
		}
	}

	switch profile.Mode {
	case "union":
		if err := u.createUnionRemote(profile.Name, mounts); err != nil {
			return err
		}
	case "combine":
		if err := u.createCombineRemote(profile.Name, mounts); err != nil {
			return err
		}
	default:
		return errors.New("unsupported rclone mode")
	}
	return u.updateManagedRemotes(profile, desiredRemotes)
}

func (u *RcloneUseCase) applyProfileConfigWithoutPassword(profile *entity.RcloneProfile, mounts []entity.MountPoint) error {
	managedRemotes, _ := parseStringSlice(profile.ManagedRemotes)
	desiredRemotes := u.expectedRemoteNames(profile, mounts)
	if err := u.cleanupManagedRemotes(managedRemotes, desiredRemotes); err != nil {
		return err
	}

	existingRemotes, err := u.listRemotes()
	if err != nil {
		return err
	}

	if profile.Mode == "ordinary" {
		mount := mounts[0]
		if err := u.updateExistingWebDAVRemoteWithoutPassword(existingRemotes, profile.Name, mount.ID, profile.Username); err != nil {
			return err
		}
		return u.updateManagedRemotes(profile, desiredRemotes)
	}

	for _, mount := range mounts {
		underlyingRemote := u.underlyingRemoteName(profile.Name, mount.ID)
		if err := u.updateExistingWebDAVRemoteWithoutPassword(existingRemotes, underlyingRemote, mount.ID, profile.Username); err != nil {
			return err
		}
	}

	switch profile.Mode {
	case "union":
		if err := u.createUnionRemote(profile.Name, mounts); err != nil {
			return err
		}
	case "combine":
		if err := u.createCombineRemote(profile.Name, mounts); err != nil {
			return err
		}
	default:
		return errors.New("unsupported rclone mode")
	}
	return u.updateManagedRemotes(profile, desiredRemotes)
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

func (u *RcloneUseCase) updateWebDAVRemote(remoteName string, mountID uint, username string) error {
	args := []string{
		"config", "update", remoteName,
		"url", u.mountWebDAVURL(mountID),
		"vendor", "other",
		"user", username,
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

func (u *RcloneUseCase) updateExistingWebDAVRemoteWithoutPassword(existingRemotes map[string]struct{}, remoteName string, mountID uint, username string) error {
	if _, ok := existingRemotes[remoteName]; !ok {
		return fmt.Errorf("rclone remote %q no longer exists; please edit the profile and re-enter the password once", remoteName)
	}
	return u.updateWebDAVRemote(remoteName, mountID, username)
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

func (u *RcloneUseCase) currentOpenListScope() string {
	return config.NormalizeBaseURLScope(u.config.OpenList.BaseURL)
}

func (u *RcloneUseCase) decryptProfilePassword(profile *entity.RcloneProfile, mounts []entity.MountPoint, ciphertext string) (string, bool, error) {
	if strings.TrimSpace(ciphertext) == "" {
		return "", false, errors.New("saved rclone password is empty")
	}
	password, err := u.secretCodec.Decrypt(ciphertext)
	if err == nil {
		return password, false, nil
	}
	for _, codec := range u.legacyCodecs {
		password, legacyErr := codec.Decrypt(ciphertext)
		if legacyErr == nil {
			return password, true, nil
		}
	}
	legacyPorts, portsErr := u.legacyRemotePorts(profile, mounts)
	if portsErr == nil {
		for _, port := range legacyPorts {
			if strings.TrimSpace(port) == "" || strings.TrimSpace(port) == strings.TrimSpace(u.config.App.Port) {
				continue
			}
			codec := tool.NewSecretCodec(strings.Join([]string{u.config.App.Name, u.config.DB.Path, port}, "|"))
			password, legacyErr := codec.Decrypt(ciphertext)
			if legacyErr == nil {
				return password, true, nil
			}
		}
	}
	return "", false, errors.New(u.passwordRecoveryHint())
}

func (u *RcloneUseCase) passwordRecoveryHint() string {
	return "saved rclone password can no longer be decrypted after the app port/source changed; the existing remote will be refreshed when possible, otherwise please edit the profile and re-enter the password once"
}

func (u *RcloneUseCase) listRemotes() (map[string]struct{}, error) {
	output, err := u.runRcloneCommand(10*time.Second, "listremotes")
	if err != nil {
		return nil, err
	}
	remotes := map[string]struct{}{}
	for _, line := range strings.Split(output, "\n") {
		name := strings.TrimSpace(strings.TrimSuffix(line, ":"))
		if name == "" {
			continue
		}
		remotes[name] = struct{}{}
	}
	return remotes, nil
}

func (u *RcloneUseCase) legacyRemotePorts(profile *entity.RcloneProfile, mounts []entity.MountPoint) ([]string, error) {
	existingRemotes, err := u.listRemotes()
	if err != nil {
		return nil, err
	}
	candidates := expectedWebDAVRemoteNames(profile, mounts)
	ports := make([]string, 0, len(candidates))
	seen := map[string]struct{}{}
	for _, remoteName := range candidates {
		if _, ok := existingRemotes[remoteName]; !ok {
			continue
		}
		port, portErr := u.remoteWebDAVPort(remoteName)
		if portErr != nil || strings.TrimSpace(port) == "" {
			continue
		}
		if _, ok := seen[port]; ok {
			continue
		}
		seen[port] = struct{}{}
		ports = append(ports, port)
	}
	return ports, nil
}

func (u *RcloneUseCase) remoteWebDAVPort(remoteName string) (string, error) {
	output, err := u.runRcloneCommand(10*time.Second, "config", "show", remoteName)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(strings.ToLower(trimmed), "url") {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}
		rawURL := strings.TrimSpace(parts[1])
		if rawURL == "" {
			continue
		}
		parsed, parseErr := url.Parse(rawURL)
		if parseErr != nil {
			continue
		}
		if parsed.Port() != "" {
			return parsed.Port(), nil
		}
	}
	return "", errors.New("webdav remote url port not found")
}

func (u *RcloneUseCase) startMountProcess(profile *entity.RcloneProfile) (int, string, error) {
	rclonePath := resolvedRclonePath(u.config.Rclone.Path)
	rcAddr := u.mountRCAddr(profile.ID)
	args := []string{"mount", profile.Name + ":", profile.TargetPath, "--vfs-cache-mode", "full", "--rc", "--rc-no-auth", "--rc-addr", rcAddr}
	if runtime.GOOS == "windows" {
		args = append(args, "--links")
	}

	if runtime.GOOS != "windows" {
		args = append(args, "--daemon")
	}
	cmd := exec.Command(rclonePath, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return 0, "", err
	}
	pid := 0
	if cmd.Process != nil {
		pid = cmd.Process.Pid
	}
	return pid, rcAddr, nil
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

func (u *RcloneUseCase) stopMountedProfile(profile *entity.RcloneProfile) error {
	stopErrors := make([]string, 0, 2)

	if strings.TrimSpace(profile.MountRCAddr) != "" {
		if _, err := u.runRcloneCommand(10*time.Second, "rc", "--rc-addr", profile.MountRCAddr, "core/quit"); err != nil {
			stopErrors = append(stopErrors, err.Error())
		} else {
			time.Sleep(800 * time.Millisecond)
		}
	}

	if profile.MountPID > 0 {
		if err := killProcess(profile.MountPID); err != nil {
			stopErrors = append(stopErrors, err.Error())
		}
	}

	profile.IsMounted = false
	profile.MountPID = 0
	profile.MountRCAddr = ""
	if len(stopErrors) > 0 {
		profile.LastError = strings.Join(stopErrors, " | ")
	} else {
		profile.LastError = ""
	}
	return nil
}

func (u *RcloneUseCase) syncProfiles() error {
	profiles, err := u.repo.List(u.currentOpenListScope())
	if err != nil {
		return err
	}
	for i := range profiles {
		profile := profiles[i]
		if err := u.reconcileProfileRuntime(&profile); err != nil {
			return err
		}
	}
	return nil
}

func (u *RcloneUseCase) reconcileProfileRuntime(profile *entity.RcloneProfile) error {
	mountIDs, err := parseMountIDs(profile.MountIDs)
	if err != nil {
		return nil
	}
	validMountIDs := make([]uint, 0, len(mountIDs))
	for _, mountID := range mountIDs {
		mount, getErr := u.mountRepo.GetMountPoint(mountID)
		if getErr != nil {
			continue
		}
		if _, scopeErr := u.providerRepo.GetProviderAccountByOpenList(mount.ProviderAccountID, config.NormalizeBaseURLScope(u.config.OpenList.BaseURL)); scopeErr != nil {
			continue
		}
		validMountIDs = append(validMountIDs, mountID)
	}

	changed := false
	if len(validMountIDs) != len(mountIDs) {
		changed = true
		if err := u.reconcileProfileAfterMountIDs(profile, validMountIDs, mountIDs); err != nil {
			return err
		}
	}

	if profile.IsMounted && strings.TrimSpace(profile.MountRCAddr) != "" {
		if _, err := u.runRcloneCommand(5*time.Second, "rc", "--rc-addr", profile.MountRCAddr, "core/version"); err != nil {
			profile.IsMounted = false
			profile.MountPID = 0
			profile.MountRCAddr = ""
			changed = true
		}
	}

	if changed {
		return u.repo.Update(profile)
	}
	return nil
}

func (u *RcloneUseCase) reconcileProfileAfterMountChange(profile *entity.RcloneProfile, removedMountID uint) error {
	mountIDs, err := parseMountIDs(profile.MountIDs)
	if err != nil {
		return nil
	}
	previousMountIDs := append([]uint(nil), mountIDs...)
	next := make([]uint, 0, len(mountIDs))
	for _, mountID := range mountIDs {
		if mountID != removedMountID {
			next = append(next, mountID)
		}
	}
	return u.reconcileProfileAfterMountIDs(profile, next, previousMountIDs)
}

func (u *RcloneUseCase) reconcileProfileAfterMountIDs(profile *entity.RcloneProfile, mountIDs []uint, previousMountIDs []uint) error {
	if len(mountIDs) == 0 {
		if err := u.stopMountedProfile(profile); err != nil {
			return err
		}
		managedRemotes, _ := parseStringSlice(profile.ManagedRemotes)
		if len(managedRemotes) == 0 {
			managedRemotes = u.legacyRemoteNamesFromIDs(profile.Name, profile.Mode, previousMountIDs)
		}
		if err := u.cleanupManagedRemotes(managedRemotes, nil); err != nil {
			return err
		}
		return u.repo.Delete(profile.ID, u.currentOpenListScope())
	}

	if len(mountIDs) == 1 {
		profile.Mode = "ordinary"
	} else if profile.Mode == "ordinary" {
		profile.Mode = "union"
	}

	mountIDs = uniqueUintSlice(mountIDs)
	rawMountIDs, err := json.Marshal(mountIDs)
	if err != nil {
		return err
	}
	profile.MountIDs = string(rawMountIDs)

	mounts, _ := u.resolveMounts(profile)
	desiredRemotes := u.expectedRemoteNames(profile, mounts)
	managedRemotes, _ := parseStringSlice(profile.ManagedRemotes)
	if len(managedRemotes) == 0 {
		managedRemotes = u.legacyRemoteNamesFromIDs(profile.Name, profile.Mode, previousMountIDs)
	}
	if err := u.cleanupManagedRemotes(managedRemotes, desiredRemotes); err != nil {
		return err
	}
	profile.ManagedRemotes = mustMarshalStringSlice(desiredRemotes)
	profile.LastError = "rclone profile updated after mount changes"
	return u.repo.Update(profile)
}

func (u *RcloneUseCase) updateManagedRemotes(profile *entity.RcloneProfile, remotes []string) error {
	profile.ManagedRemotes = mustMarshalStringSlice(remotes)
	return u.repo.Update(profile)
}

func (u *RcloneUseCase) cleanupManagedRemotes(existing []string, keep []string) error {
	if len(existing) == 0 {
		return nil
	}
	keepSet := map[string]struct{}{}
	for _, remote := range keep {
		keepSet[remote] = struct{}{}
	}
	for _, remote := range existing {
		if _, ok := keepSet[remote]; ok {
			continue
		}
		if err := u.deleteRemote(remote); err != nil {
			return err
		}
	}
	return nil
}

func (u *RcloneUseCase) expectedRemoteNames(profile *entity.RcloneProfile, mounts []entity.MountPoint) []string {
	names := make([]string, 0, len(mounts)+1)
	if profile.Mode == "ordinary" {
		return []string{profile.Name}
	}
	for _, mount := range mounts {
		names = append(names, u.underlyingRemoteName(profile.Name, mount.ID))
	}
	names = append(names, profile.Name)
	sort.Strings(names)
	return names
}

func (u *RcloneUseCase) legacyRemoteNamesFromIDs(profileName string, mode string, mountIDs []uint) []string {
	if mode == "ordinary" {
		return []string{profileName}
	}
	names := make([]string, 0, len(mountIDs)+1)
	for _, mountID := range mountIDs {
		names = append(names, u.underlyingRemoteName(profileName, mountID))
	}
	names = append(names, profileName)
	sort.Strings(names)
	return names
}

func (u *RcloneUseCase) mountRCAddr(profileID uint) string {
	port := 56000 + int(profileID%4000)
	return fmt.Sprintf("127.0.0.1:%d", port)
}

func (u *RcloneUseCase) mountWebDAVURL(mountID uint) string {
	return fmt.Sprintf("http://127.0.0.1:%s/api/v1/webdav/mounts/%d", u.config.App.Port, mountID)
}

func (u *RcloneUseCase) underlyingRemoteName(profileName string, mountID uint) string {
	return fmt.Sprintf("%s__mount_%d", profileName, mountID)
}

func expectedWebDAVRemoteNames(profile *entity.RcloneProfile, mounts []entity.MountPoint) []string {
	if profile.Mode == "ordinary" {
		return []string{profile.Name}
	}
	names := make([]string, 0, len(mounts))
	for _, mount := range mounts {
		names = append(names, fmt.Sprintf("%s__mount_%d", profile.Name, mount.ID))
	}
	return names
}

func parseMountIDs(raw string) ([]uint, error) {
	var ids []uint
	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func parseStringSlice(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return []string{}, nil
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, err
	}
	return values, nil
}

func mustMarshalStringSlice(values []string) string {
	data, err := json.Marshal(values)
	if err != nil {
		return "[]"
	}
	return string(data)
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
		return " --links --rc --rc-no-auth --rc-addr 127.0.0.1:<port>"
	}
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		return " --daemon --rc --rc-no-auth --rc-addr 127.0.0.1:<port>"
	}
	return ""
}

func killProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if process == nil {
		return nil
	}
	if err := process.Kill(); err != nil && !strings.Contains(strings.ToLower(err.Error()), "finished") {
		return err
	}
	return nil
}
