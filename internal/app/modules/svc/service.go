package svc

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"assistant/internal/bootstrap/psl"
	"assistant/internal/platform"
	"assistant/pkg/dwmblocknotify"
	"assistant/pkg/utils"

	"github.com/sirupsen/logrus"
	"golang.design/x/clipboard"
)

type Service struct {
	logger *logrus.Logger
}

func NewService() *Service {
	return &Service{logger: psl.GetLogger()}
}

var clipboardOnce sync.Once

func initClipboard() error {
	var err error
	clipboardOnce.Do(func() {
		err = clipboard.Init()
	})
	return err
}

func (s *Service) readClipboard() (string, error) {
	if err := initClipboard(); err != nil {
		return "", fmt.Errorf("init clipboard: %w", err)
	}
	data := clipboard.Read(clipboard.FmtText)
	return string(data), nil
}

func (s *Service) writeClipboard(content string) error {
	if err := initClipboard(); err != nil {
		return fmt.Errorf("init clipboard: %w", err)
	}
	clipboard.Write(clipboard.FmtText, []byte(content))
	return nil
}

// pushClipboard writes content to the clipboard, surfaces a notification, and
// returns a copy of the value for the HTTP response.
func (s *Service) pushClipboard(value, summary string) (string, error) {
	if err := s.writeClipboard(value); err != nil {
		s.logger.WithError(err).Warn("write clipboard failed")
		return value, err
	}
	dwmblocknotify.PUT(summary, 2*time.Second)
	return value, nil
}

// copyToClipboardWithNotify is the fire-and-forget variant of pushClipboard.
func (s *Service) copyToClipboardWithNotify(value, summary string) {
	if err := s.writeClipboard(value); err != nil {
		s.logger.WithError(err).Warn("write clipboard failed")
		return
	}
	dwmblocknotify.PUT(summary, 2*time.Second)
}

// writeAIOutput writes generated content under the configured AI output dir.
func (s *Service) writeAIOutput(rel, content string) error {
	dir := psl.GetConfig().Settings.DirAIOutput
	if strings.TrimSpace(dir) == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home dir: %w", err)
		}
		dir = filepath.Join(home, "git", "test")
	}
	out := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	if err := os.WriteFile(out, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

// captureConfig builds the ffmpeg device config from settings.
func captureConfig() platform.CaptureConfig {
	cfg := psl.GetConfig().Settings
	return platform.CaptureConfig{
		ScreenDevice: cfg.CaptureScreenDevice,
		AudioDevice:  cfg.CaptureAudioDevice,
		VideoDevice:  cfg.CaptureVideoDevice,
	}
}

func (s *Service) Launch(command string) error {
	return utils.StartScript("bash", command)
}

func (s *Service) Toggle(cmd, match string) (string, error) {
	if match == "" {
		return "", fmt.Errorf("match cannot be empty")
	}
	s.logger.Info(fmt.Sprintf("toggle %s", cmd))
	if utils.ProcessRunning(match) {
		s.logger.Info("running")
		if err := utils.KillProcess(match); err != nil {
			s.logger.Info("killed")
			return "", fmt.Errorf("kill %s: %w", match, err)
		}
		return "killed", nil
	}
	if err := utils.StartScript("bash", cmd); err != nil {
		return "", fmt.Errorf("launch %s: %w", cmd, err)
	}
	return "launched", nil
}

func (s *Service) ToggleTTYClock() (string, error) {
	term := psl.GetConfig().Settings.DefaultTerminal
	cmd, err := platform.Get().FloatingTerminalCommand(term, "tty-clock -s")
	if err != nil {
		return "", err
	}
	return s.Toggle(cmd, "tty-clock")
}

func (s *Service) ToggleMusic() (string, error) {
	if utils.ProcessRunning("ncmpcpp") || utils.ProcessRunning("cava") {
		if err := utils.KillProcess("ncmpcpp"); err != nil {
			return "", fmt.Errorf("kill ncmpcpp: %w", err)
		}
		if err := utils.KillProcess("cava"); err != nil {
			return "", fmt.Errorf("kill cava: %w", err)
		}
		return "killed", nil
	}
	term := psl.GetConfig().Settings.DefaultTerminal
	ncmpcppCmd, err := platform.Get().TerminalCommand(term, "ncmpcpp")
	if err != nil {
		return "", err
	}
	if err := utils.StartScript("bash", ncmpcppCmd); err != nil {
		return "", fmt.Errorf("launch ncmpcpp: %w", err)
	}
	cavaCmd, err := platform.Get().TerminalCommand(term, "cava")
	if err != nil {
		return "", err
	}
	if err := utils.StartScript("bash", cavaCmd); err != nil {
		return "", fmt.Errorf("launch cava: %w", err)
	}
	return "launched", nil
}

func (s *Service) ToggleRecShow() (string, error) {
	term := psl.GetConfig().Settings.DefaultTerminal
	script := platform.CommandLine(platform.Get().PreviewWebcamCmd(captureConfig()))
	cmd, err := platform.Get().TerminalCommand(term, script)
	if err != nil {
		return "", err
	}
	return s.Toggle(cmd, "ffplay")
}

func (s *Service) ToggleRecAudio() (string, error) {
	if utils.ProcessRunning("ffmpeg") {
		if err := utils.KillProcess("ffmpeg"); err != nil {
			return "", fmt.Errorf("kill ffmpeg: %w", err)
		}
		return "killed", nil
	}
	homeDir, _ := os.UserHomeDir()
	now := time.Now().Format("2006-01-02-15-04-05")
	term := psl.GetConfig().Settings.DefaultTerminal
	filename := path.Join(homeDir, fmt.Sprintf("Videos/rec-audio-%s.flac", now))
	script := platform.CommandLine(platform.Get().CaptureAudioArgs(filename, captureConfig()))
	cmd, err := platform.Get().ScratchpadTerminalCommand(term, script)
	if err != nil {
		return "", err
	}
	return s.Toggle(cmd, "ffmpeg")
}

func (s *Service) ToggleRecScreen() (string, error) {
	if utils.ProcessRunning("ffmpeg") {
		if err := utils.KillProcess("ffmpeg"); err != nil {
			return "", fmt.Errorf("kill ffmpeg: %w", err)
		}
		return "killed", nil
	}
	homeDir, _ := os.UserHomeDir()
	now := time.Now().Format("2006-01-02-15-04-05")
	term := psl.GetConfig().Settings.DefaultTerminal
	filename := path.Join(homeDir, fmt.Sprintf("Videos/rec-screen-%s.mkv", now))
	w, h := platform.Get().ScreenSize()
	script := platform.CommandLine(platform.Get().CaptureScreenArgs(filename, w, h, captureConfig()))
	cmd, err := platform.Get().ScratchpadTerminalCommand(term, script)
	if err != nil {
		return "", err
	}
	return s.Toggle(cmd, "ffmpeg")
}

func (s *Service) ToggleRecWebcam() (string, error) {
	if utils.ProcessRunning("ffmpeg") {
		if err := utils.KillProcess("ffmpeg"); err != nil {
			return "", fmt.Errorf("kill ffmpeg: %w", err)
		}
		return "killed", nil
	}
	homeDir, _ := os.UserHomeDir()
	now := time.Now().Format("2006-01-02-15-04-05")
	term := psl.GetConfig().Settings.DefaultTerminal
	filename := path.Join(homeDir, fmt.Sprintf("Videos/rec-webcam-%s.mp4", now))
	script := platform.CommandLine(platform.Get().CaptureWebcamArgs(filename, captureConfig()))
	cmd, err := platform.Get().ScratchpadTerminalCommand(term, script)
	if err != nil {
		return "", err
	}
	return s.Toggle(cmd, "ffmpeg")
}
