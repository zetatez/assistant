package psl

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
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
		// One-shot commands (e.g. xset) are executed once per service run; they
		// are not processes to keep alive, so they must not be re-checked with
		// pgrep on every tick.
		oneshotDone := make(map[string]bool)
		// Names whose command is not installed: report once instead of trying
		// (and failing) every tick forever.
		missing := make(map[string]bool)
		for {
			for _, p := range procs {
				if p.OneShot && oneshotDone[p.Name] {
					continue
				}
				if !p.OneShot && isRunning(p.Name) {
					continue
				}
				if p.Precursor != "" && !isRunning(p.Precursor) {
					continue
				}
				if missing[p.Name] {
					continue
				}
				if bin := commandBinary(p.Command); bin != "" {
					if _, err := exec.LookPath(bin); err != nil {
						GetLogger().WithField("proc", p.Name).Warnf("proc command not found, skipping: %s", bin)
						missing[p.Name] = true
						continue
					}
				}
				command := p.Command
				if !p.OneShot {
					command += " &"
				}
				cmd := exec.CommandContext(ctx, "bash", "-c", command)
				if p.OneShot {
					if err := cmd.Run(); err != nil {
						GetLogger().WithError(err).WithField("proc", p.Name).Warn("run oneshot proc failed")
					} else {
						GetLogger().WithField("proc", p.Name).Info("ran oneshot proc")
					}
					oneshotDone[p.Name] = true
					continue
				}
				if err := cmd.Start(); err != nil {
					GetLogger().WithError(err).WithField("proc", p.Name).Warn("start proc failed")
					continue
				}
				GetLogger().WithField("proc", p.Name).Info("started proc")
				go func() { _ = cmd.Wait() }()
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

// commandBinary returns the executable of a shell command (first field), used
// to detect commands that are not installed.
func commandBinary(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
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
