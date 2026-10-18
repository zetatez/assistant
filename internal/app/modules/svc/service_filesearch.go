package svc

import (
	"fmt"
	"os"
	"strings"

	"assistant/internal/bootstrap/psl"
	"assistant/internal/platform"
	"assistant/pkg/utils"
)

const (
	tmplFileSearch = `fd --type f \
  --hidden \
  --exclude .git \
  --exclude .cache \
  --exclude .local \
  . "%s" \
| fzf \
  --prompt="search file>" \
  --preview "lazy -o view -f {}" \
  --select-1 \
  --exit-0 \
  --print0 \
| xargs -0 -o lazy -o open -f`

	tmplFileSearchContent = `cd "%s" && \
RG_PREFIX="rg --column --line-number --no-heading --color=always --smart-case"
matched=$( \
	FZF_DEFAULT_COMMAND="$RG_PREFIX " \
	fzf \
  --prompt="search file content> " \
  --preview "bat --style=full --color=always --highlight-line {2} {1}" \
  --bind "change:reload:$RG_PREFIX {q} || true" \
  --ansi \
	--disabled \
	--query "" \
  --height=100%% \
	--layout=reverse \
	--delimiter : \
)
filepath=$(echo "$matched" | awk -F: "{print \$1}")
rowno=$(echo "$matched" | awk -F: "{print \$2}")
if [ -n "$filepath" ]; then
  nvim +"$rowno" "$filepath"
fi`

	tmplFileSearchBook = `fd --type f \
  --extension pdf \
  --extension epub \
  --extension djvu \
  --extension mobi \
  --hidden \
  --exclude .git \
  --exclude .cache \
  --exclude .local \
  . "%s" \
| fzf \
  --prompt="search book>" \
  --preview "lazy -o view -f {}" \
  --select-1 \
  --exit-0 \
  --print0 \
| xargs -0 -o lazy -o open -f`

	tmplFileSearchMedia = `fd --type f \
  --extension jpg \
  --extension jpeg \
  --extension png \
  --extension gif \
  --extension bmp \
  --extension tiff \
  --extension avi \
  --extension flac \
  --extension mkv \
  --extension mp3 \
  --extension mp4 \
  --hidden \
  --exclude .git \
  --exclude .cache \
  --exclude .local \
  . "%s" \
| fzf \
  --prompt="search media>" \
  --preview "lazy -o view -f {}" \
  --select-1 \
  --exit-0 \
  --print0 \
| xargs -0 -o lazy -o open -f`

	tmplFileSearchWiki = `fd --type f \
  --extension md \
  --hidden \
  --exclude .git \
  --exclude .cache \
  --exclude .local \
  . "%s" \
| fzf \
  --prompt="search wiki>" \
  --preview "lazy -o view -f {}" \
  --select-1 \
  --exit-0 \
  --print0 \
| xargs -0 -o lazy -o open -f`

	tmplFileSearchExec = `fd --type x \
  --extension sh \
  --extension py \
  --extension jl \
  --hidden \
  --exclude .git \
  --exclude .cache \
  --exclude .local \
  . "%s" \
| fzf \
  --prompt="lazy exec search file>" \
  --preview "lazy -o view -f {}" \
  --select-1 \
  --exit-0 \
  --print0 \
| xargs -0 -o lazy -o exec -f`
)

func homeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

// runInTerminal starts script inside the configured terminal.
func runInTerminal(script string) error {
	term := psl.GetConfig().Settings.DefaultTerminal
	cmd, err := platform.Get().TerminalCommand(term, script)
	if err != nil {
		return err
	}
	return utils.StartScript("bash", cmd)
}

func (s *Service) SearchWeb(query string) error {
	url := "https://www.google.com/search?q=" + strings.ReplaceAll(query, " ", "+")
	return s.OpenURL("chrome", url)
}

func (s *Service) FileSearch(dir string) error {
	if dir == "" {
		dir = homeDir()
	}
	return runInTerminal(fmt.Sprintf(tmplFileSearch, dir))
}

func (s *Service) FileSearchContent(dir string) error {
	if dir == "" {
		dir = homeDir()
	}
	return runInTerminal(fmt.Sprintf(tmplFileSearchContent, dir))
}

func (s *Service) FileSearchBook(dir string) error {
	if dir == "" {
		dir = homeDir()
	}
	return runInTerminal(fmt.Sprintf(tmplFileSearchBook, dir))
}

func (s *Service) FileSearchMedia(dir string) error {
	if dir == "" {
		dir = homeDir()
	}
	return runInTerminal(fmt.Sprintf(tmplFileSearchMedia, dir))
}

func (s *Service) FileSearchWiki(dir string) error {
	if dir == "" {
		dir = homeDir()
	}
	return runInTerminal(fmt.Sprintf(tmplFileSearchWiki, dir))
}

func (s *Service) FileSearchExec(dir string) error {
	if dir == "" {
		dir = homeDir()
	}
	return runInTerminal(fmt.Sprintf(tmplFileSearchExec, dir))
}

func (s *Service) OpenImages(dir string) error {
	if dir == "" {
		dir = homeDir()
	}
	return platform.Get().OpenImages(dir)
}
