package psl

import (
	"context"
	"os/exec"
	"path/filepath"
	"time"

	"assistant/internal/platform"
)

func StartBackgroundTasks(ctx context.Context) {
	cfg := GetConfig()
	if cfg.Background.Enabled {
		startDaemon(ctx, cfg.Background.Procs)
	}
	startWallpaper(ctx)
}

func startDaemon(ctx context.Context, procs []BackgroundProc) {
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			for _, p := range procs {
				if !isRunning(p.Name) {
					if p.Precursor != "" && !isRunning(p.Precursor) {
						continue
					}
					cmd := exec.CommandContext(ctx, "bash", "-c", p.Command+" &")
					if err := cmd.Start(); err != nil {
						GetLogger().WithError(err).WithField("proc", p.Name).Warn("start proc failed")
						continue
					}
					go func() { _ = cmd.Wait() }()
				}
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func isRunning(name string) bool {
	if name == "" {
		return false
	}
	return exec.Command("pgrep", "-x", name).Run() == nil
}

func startWallpaper(ctx context.Context) {
	go func() {
		dir := GetConfig().Settings.DirWallpaper
		if dir == "" {
			return
		}
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			for _, pic := range wallpaperFiles(dir) {
				select {
				case <-ctx.Done():
					return
				default:
				}
				if err := platform.Get().SetWallpaper(pic); err != nil {
					GetLogger().WithError(err).WithField("picture", pic).Warn("set wallpaper failed")
					break
				}
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
				}
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

// wallpaperFiles returns the supported image files in dir, newest behaviour
// compatible with the previous *.JPG-only glob but covering common extensions.
func wallpaperFiles(dir string) []string {
	var files []string
	for _, pattern := range []string{"*.jpg", "*.jpeg", "*.png", "*.JPG", "*.JPEG", "*.PNG"} {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			continue
		}
		files = append(files, matches...)
	}
	return files
}
