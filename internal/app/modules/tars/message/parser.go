package message

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Parser struct {
	logger Logger
}

type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
}

type UserMessage struct {
	Text      string
	ImageKey  string
	FileKey   string
	FileName  string
	MimeType  string
	Supported bool
	Skip      bool
}

func NewParser(logger Logger) *Parser {
	return &Parser{logger: logger}
}

var excludedExts = []string{
	".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz",
	".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".webm",
	".mp3", ".wav", ".flac", ".aac", ".ogg",
	".exe", ".dll", ".so", ".app",
}

var textExts = []string{
	".txt", ".md", ".json", ".yaml", ".yml", ".xml", ".html", ".htm",
	".css", ".js", ".ts", ".py", ".go", ".java", ".c", ".cpp", ".h",
	".sh", ".bash", ".zsh", ".ps1", ".bat", ".cmd", ".sql", ".log",
	".conf", ".config", ".ini", ".toml",
}

func (p *Parser) Parse(msgType string, content string) UserMessage {
	switch msgType {
	case "text":
		return p.parseText(content)
	case "image":
		return p.parseImage(content)
	case "audio":
		return p.parseAudio(content)
	case "video":
		return UserMessage{Supported: false, Skip: true}
	case "file":
		return p.parseFile(content)
	case "post":
		return p.parsePost(content)
	case "share_chat", "share_user", "redpacket", "sticker", "emotion":
		return UserMessage{Supported: false, Skip: true}
	default:
		p.logger.Infof("unsupported message type: %s", msgType)
		return UserMessage{Supported: false, Skip: true}
	}
}

func (p *Parser) parseText(content string) UserMessage {
	var text struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(content), &text); err != nil {
		p.logger.Errorf("failed to parse text content: %v", err)
		return UserMessage{Supported: false}
	}
	return UserMessage{
		Text:      strings.TrimSpace(text.Text),
		Supported: true,
	}
}

func (p *Parser) parseImage(content string) UserMessage {
	var img struct {
		ImageKey string `json:"image_key"`
		Height   int    `json:"height"`
		Width    int    `json:"width"`
	}
	if err := json.Unmarshal([]byte(content), &img); err != nil {
		return UserMessage{
			Text:      "[Image]",
			Supported: true,
		}
	}
	desc := "[Image]"
	if img.Height > 0 && img.Width > 0 {
		desc = fmt.Sprintf("[Image: %dx%d]", img.Width, img.Height)
	}
	return UserMessage{
		Text:      desc,
		ImageKey:  img.ImageKey,
		Supported: true,
	}
}

func (p *Parser) parseAudio(content string) UserMessage {
	var audio struct {
		Duration int    `json:"duration"`
		FileKey  string `json:"file_key"`
	}
	if err := json.Unmarshal([]byte(content), &audio); err != nil {
		return UserMessage{Supported: false, Skip: true}
	}
	durationSec := audio.Duration / 1000
	return UserMessage{
		Text:      fmt.Sprintf("[Audio message: %d seconds, key=%s] This audio cannot be parsed directly.", durationSec, audio.FileKey),
		Supported: true,
	}
}

func (p *Parser) parseFile(content string) UserMessage {
	var file struct {
		FileName string `json:"file_name"`
		FileSize int64  `json:"file_size"`
		MimeType string `json:"mime_type"`
		FileKey  string `json:"file_key"`
	}
	if err := json.Unmarshal([]byte(content), &file); err != nil {
		return UserMessage{
			Text:      "[File]",
			Supported: true,
		}
	}
	if isExcludedFileType(file.FileName) {
		p.logger.Infof("skipping excluded file type: %s", file.FileName)
		return UserMessage{Supported: false, Skip: true}
	}
	sizeStr := formatFileSize(file.FileSize)
	if isTextFileType(file.FileName) {
		return UserMessage{
			Text:      fmt.Sprintf("[Text file: %s (%s), key=%s]", file.FileName, sizeStr, file.FileKey),
			FileKey:   file.FileKey,
			FileName:  file.FileName,
			MimeType:  file.MimeType,
			Supported: true,
		}
	}
	return UserMessage{
		Text:      fmt.Sprintf("[File: %s (%s), key=%s] This file type cannot be parsed directly.", file.FileName, sizeStr, file.FileKey),
		Supported: true,
	}
}

func (p *Parser) parsePost(content string) UserMessage {
	var post struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(content), &post); err != nil {
		return UserMessage{
			Text:      "[Rich text post]",
			Supported: true,
		}
	}
	text := post.Title
	if post.Content != "" {
		text += "\n" + post.Content
	}
	return UserMessage{
		Text:      strings.TrimSpace(text),
		Supported: true,
	}
}

func isExcludedFileType(filename string) bool {
	lower := strings.ToLower(filename)
	for _, ext := range excludedExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

func isTextFileType(filename string) bool {
	lower := strings.ToLower(filename)
	for _, ext := range textExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
