package psl

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"
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
	return exec.Command("pgrep", "-x", name).Run() == nil
}

func startWallpaper(ctx context.Context) {
	go func() {
		dir := GetConfig().Settings.DirWallpaper
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			files, _ := filepath.Glob(filepath.Join(dir, "*.JPG"))
			for _, pic := range files {
				select {
				case <-ctx.Done():
					return
				default:
				}
				cmd := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("feh --bg-fill '%s'", pic))
				_ = cmd.Run()
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
