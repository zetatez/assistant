package svc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"assistant/internal/bootstrap/psl"
	"assistant/internal/platform"
	"assistant/pkg/utils"
)

func (s *Service) GitLogShow(dir string) error {
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}

	repo, err := utils.OpenGitRepo(context.Background(), dir)
	if err != nil {
		return fmt.Errorf("open git repo: %w", err)
	}

	script := fmt.Sprintf(`cd %s && sha=$(git log --pretty=oneline | fzf --prompt='git show>' --ansi --preview 'git show --color=always {1}' --select-1 --exit-0 | awk '{print $1}')
[ -n "$sha" ] && git show "$sha"`, platform.Quote(repo.Root))

	term := psl.GetConfig().Settings.DefaultTerminal
	cmd, err := platform.Get().TerminalCommand(term, script)
	if err != nil {
		return fmt.Errorf("launch git show: %w", err)
	}
	if err := utils.StartScript("bash", cmd); err != nil {
		return fmt.Errorf("launch git show: %w", err)
	}
	return nil
}
