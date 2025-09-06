package utils

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
)

// -------------------- 环境变量 --------------------
func GetEnv(key, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultVal
}

func SetEnv(key, value string) error {
	return os.Setenv(key, value)
}

// -------------------- 系统信息 --------------------
func GetOSType() string {
	return runtime.GOOS
}

func GetArch() string {
	return runtime.GOARCH
}

func GetHostname() (string, error) {
	return os.Hostname()
}

func GetCurrentUser() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.Username, nil
}

// -------------------- 文件/目录 --------------------
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func IsFile(path string) (isFile bool) {
	return !IsDir(path)
}

func IsDirExists(path string) (exist bool) {
	if Exists(path) && IsDir(path) {
		return true
	}
	return false
}

func IsFileExists(path string) (exist bool) {
	if Exists(path) && IsFile(path) {
		return true
	}
	return false
}

func MkdirAll(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func Remove(path string) error {
	return os.RemoveAll(path)
}

func GetAbsPath(path string) (string, error) {
	return filepath.Abs(path)
}

// -------------------- 进程 --------------------
func GetPID() int {
	return os.Getpid()
}

// -------------------- 系统控制 --------------------
func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Shutdown 关机
func Shutdown() error {
	impls := map[string]func() error{
		"linux":  func() error { return runCmd("systemctl", "poweroff") },
		"darwin": func() error { return runCmd("osascript", "-e", `tell application "System Events" to shut down`) },
	}
	if fn, ok := impls[runtime.GOOS]; ok {
		return fn()
	}
	return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

// Reboot 重启
func Reboot() error {
	impls := map[string]func() error{
		"linux":  func() error { return runCmd("systemctl", "reboot") },
		"darwin": func() error { return runCmd("osascript", "-e", `tell application "System Events" to restart`) },
	}
	if fn, ok := impls[runtime.GOOS]; ok {
		return fn()
	}
	return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

// Suspend / Sleep
func Suspend() error {
	impls := map[string]func() error{
		"linux":  func() error { return runCmd("systemctl", "suspend") },
		"darwin": func() error { return runCmd("pmset", "sleepnow") },
	}
	if fn, ok := impls[runtime.GOOS]; ok {
		return fn()
	}
	return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

// Lock 锁屏
func Lock() error {
	impls := map[string]func() error{
		"linux": func() error { return runCmd("slock") },
		"darwin": func() error {
			return runCmd("osascript", "-e", `tell application "System Events" to keystroke "q" using {control down, command down}`)
		},
	}
	if fn, ok := impls[runtime.GOOS]; ok {
		return fn()
	}
	return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

// Volume 设置音量 (0-100)
func Volume(level int) error {
	impls := map[string]func() error{
		"linux":  func() error { return runCmd("amixer", "set", "Master", fmt.Sprintf("%d%%", level)) },
		"darwin": func() error { return runCmd("osascript", "-e", fmt.Sprintf("set volume output volume %d", level)) },
	}
	if fn, ok := impls[runtime.GOOS]; ok {
		return fn()
	}
	return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

// Mute 静音
func Mute() error {
	impls := map[string]func() error{
		"linux":  func() error { return runCmd("amixer", "set", "Master", "mute") },
		"darwin": func() error { return runCmd("osascript", "-e", "set volume output muted true") },
	}
	if fn, ok := impls[runtime.GOOS]; ok {
		return fn()
	}
	return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

// Unmute 取消静音
func Unmute() error {
	impls := map[string]func() error{
		"linux":  func() error { return runCmd("amixer", "set", "Master", "unmute") },
		"darwin": func() error { return runCmd("osascript", "-e", "set volume output muted false") },
	}
	if fn, ok := impls[runtime.GOOS]; ok {
		return fn()
	}
	return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}
