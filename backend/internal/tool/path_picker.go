package tool

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type PathPickerKind string

const (
	PathPickerFile      PathPickerKind = "file"
	PathPickerDirectory PathPickerKind = "directory"
)

type PathPickerOptions struct {
	Kind        PathPickerKind
	Title       string
	CurrentPath string
	Filter      string
}

func PickLocalPath(options PathPickerOptions) (string, error) {
	switch runtime.GOOS {
	case "windows":
		return pickLocalPathWindows(options)
	case "darwin":
		return pickLocalPathDarwin(options)
	case "linux":
		return pickLocalPathLinux(options)
	default:
		return "", errors.New("path picker is not supported on this system")
	}
}

func pickLocalPathWindows(options PathPickerOptions) (string, error) {
	script := windowsFolderDialogScript()
	if options.Kind == PathPickerFile {
		script = windowsFileDialogScript()
	}

	cmd := exec.Command("powershell", "-NoProfile", "-STA", "-Command", script)
	cmd.Env = append(os.Environ(),
		"OPENBRIDGE_DIALOG_TITLE="+strings.TrimSpace(options.Title),
		"OPENBRIDGE_DIALOG_PATH="+strings.TrimSpace(options.CurrentPath),
		"OPENBRIDGE_DIALOG_FILTER="+strings.TrimSpace(options.Filter),
	)

	output, err := runDialogCommand(cmd, 2*time.Minute)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func pickLocalPathDarwin(options PathPickerOptions) (string, error) {
	title := strings.TrimSpace(options.Title)
	if title == "" {
		title = "Choose a path"
	}

	var script string
	if options.Kind == PathPickerFile {
		script = fmt.Sprintf(`POSIX path of (choose file with prompt %q)`, title)
	} else {
		script = fmt.Sprintf(`POSIX path of (choose folder with prompt %q)`, title)
	}

	cmd := exec.Command("osascript", "-e", script)
	output, err := runDialogCommand(cmd, 2*time.Minute)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func pickLocalPathLinux(options PathPickerOptions) (string, error) {
	title := strings.TrimSpace(options.Title)
	if title == "" {
		title = "Choose a path"
	}

	if zenityPath, err := exec.LookPath("zenity"); err == nil {
		args := []string{"--file-selection", "--title", title}
		if options.Kind == PathPickerDirectory {
			args = append(args, "--directory")
		}
		if options.Kind == PathPickerFile && strings.TrimSpace(options.Filter) != "" {
			args = append(args, "--file-filter", options.Filter)
		}
		cmd := exec.Command(zenityPath, args...)
		output, err := runDialogCommand(cmd, 2*time.Minute)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(output), nil
	}

	if kdialogPath, err := exec.LookPath("kdialog"); err == nil {
		args := []string{"--getopenfilename"}
		if options.Kind == PathPickerDirectory {
			args = []string{"--getexistingdirectory"}
		}
		args = append(args, strings.TrimSpace(options.CurrentPath), "--title", title)
		cmd := exec.Command(kdialogPath, args...)
		output, err := runDialogCommand(cmd, 2*time.Minute)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(output), nil
	}

	return "", errors.New("path picker requires zenity or kdialog on linux")
}

func runDialogCommand(cmd *exec.Cmd, timeout time.Duration) (string, error) {
	done := make(chan struct{})
	var output []byte
	var err error

	go func() {
		output, err = cmd.CombinedOutput()
		close(done)
	}()

	select {
	case <-done:
		text := strings.TrimSpace(string(output))
		if err != nil {
			if isDialogCancel(err, text) {
				return "", nil
			}
			if text == "" {
				return "", err
			}
			return "", errors.New(text)
		}
		return text, nil
	case <-time.After(timeout):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return "", errors.New("path picker timed out")
	}
}

func isDialogCancel(err error, text string) bool {
	if err == nil {
		return false
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}
	if exitErr.ExitCode() != 1 {
		return false
	}

	lower := strings.ToLower(strings.TrimSpace(text))
	return lower == "" ||
		strings.Contains(lower, "cancel") ||
		strings.Contains(lower, "canceled") ||
		strings.Contains(lower, "cancelled") ||
		strings.Contains(lower, "user canceled")
}

func windowsFolderDialogScript() string {
	return strings.Join([]string{
		"Add-Type -AssemblyName System.Windows.Forms",
		"$dialog = New-Object System.Windows.Forms.FolderBrowserDialog",
		"$dialog.Description = $env:OPENBRIDGE_DIALOG_TITLE",
		"$dialog.ShowNewFolderButton = $true",
		"$path = $env:OPENBRIDGE_DIALOG_PATH",
		"if ($path -and (Test-Path -LiteralPath $path)) { $dialog.SelectedPath = $path }",
		"if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) { [Console]::Out.Write($dialog.SelectedPath) }",
	}, "; ")
}

func windowsFileDialogScript() string {
	return strings.Join([]string{
		"Add-Type -AssemblyName System.Windows.Forms",
		"$dialog = New-Object System.Windows.Forms.OpenFileDialog",
		"$dialog.Title = $env:OPENBRIDGE_DIALOG_TITLE",
		"$dialog.Filter = if ($env:OPENBRIDGE_DIALOG_FILTER) { $env:OPENBRIDGE_DIALOG_FILTER } else { 'All files (*.*)|*.*' }",
		"$path = $env:OPENBRIDGE_DIALOG_PATH",
		"if ($path) {",
		"  if (Test-Path -LiteralPath $path) {",
		"    $item = Get-Item -LiteralPath $path",
		"    if ($item.PSIsContainer) { $dialog.InitialDirectory = $item.FullName } else { $dialog.InitialDirectory = Split-Path -Parent $item.FullName; $dialog.FileName = $item.Name }",
		"  } else {",
		"    $dir = Split-Path -Parent $path",
		"    if ($dir -and (Test-Path -LiteralPath $dir)) { $dialog.InitialDirectory = (Resolve-Path -LiteralPath $dir).Path }",
		"  }",
		"}",
		"if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) { [Console]::Out.Write($dialog.FileName) }",
	}, "; ")
}

func NormalizeDialogPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if len(path) == 2 && path[1] == ':' {
		return path
	}
	return filepath.Clean(path)
}
