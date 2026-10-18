package svc

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"assistant/internal/platform"
	"assistant/pkg/dwmblocknotify"
)

func (s *Service) Screenshot() (string, error) {
	data, err := platform.Get().Screenshot()
	if err != nil {
		return "", err
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	dir = filepath.Join(dir, "Pictures", "screenshots")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create screenshot dir: %w", err)
	}

	now := time.Now()
	filename := fmt.Sprintf("screenshot.%s.jpeg", now.Format("2006.01.02.15.04.05"))
	path := filepath.Join(dir, filename)

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write screenshot: %w", err)
	}
	dwmblocknotify.PUT("Screenshot saved: "+path, 2*time.Second)

	if err := s.writeClipboard(path); err != nil {
		s.logger.WithError(err).Warn("write clipboard failed")
	}
	dwmblocknotify.PUT("Clipboard: "+path, 2*time.Second)

	return path, nil
}
