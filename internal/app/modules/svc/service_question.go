package svc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"assistant/internal/bootstrap/psl"
	"assistant/pkg/dwmblocknotify"
	"assistant/pkg/llmproxy"
)

var questionSystemPrompt = `你是乐于助人的全能助手，请详细、准确、条理清晰地回答用户的问题。

要求：
1. 用 Markdown 格式输出，结构清晰（标题、列表、代码块等）
2. 默认用中文回答，除非问题本身要求使用其他语言
3. 包含必要的背景解释、具体步骤和示例
4. 只输出回答正文，不要输出任何多余内容
`

func (s *Service) SolveQuestion() error {
	text, err := s.readClipboard()
	if err != nil {
		return fmt.Errorf("read clipboard: %w", err)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("clipboard is empty")
	}
	dwmblocknotify.PUT("!...", 3*time.Second)

	client := psl.GetLLMClient()
	if client == nil {
		return fmt.Errorf("LLM client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	resp, err := llmproxy.Complete(ctx, client, text, llmproxy.WithSystemPrompt(questionSystemPrompt), llmproxy.WithTemperature(0.3))
	if err != nil {
		return fmt.Errorf("LLM request: %w", err)
	}

	result := strings.TrimSpace(resp.Content)
	result = stripThinkTags(result)
	if result == "" {
		return fmt.Errorf("LLM returned empty response")
	}

	if err := s.writeClipboard(result); err != nil {
		return fmt.Errorf("write clipboard: %w", err)
	}

	questionOutputFile := "/home/shiyi/git/test/x.md"
	if err := os.MkdirAll(filepath.Dir(questionOutputFile), 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	if err := os.WriteFile(questionOutputFile, []byte(result), 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	dwmblocknotify.PUT("!!!", 5*time.Second)
	return nil
}

func (s *Service) SolveQuestionScreenshot() error {
	dwmblocknotify.PUT("!...", 3*time.Second)
	imgBase64, err := s.screenshot()
	if err != nil {
		return fmt.Errorf("screenshot: %w", err)
	}

	client := psl.GetLLMClient()
	if client == nil {
		return fmt.Errorf("LLM client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	prompt := "请识别截图中显示的内容并回答其中的问题。"

	resp, err := llmproxy.Complete(ctx, client, prompt,
		llmproxy.WithSystemPrompt(questionSystemPrompt),
		llmproxy.WithTemperature(0.3),
		llmproxy.WithImageBase64(imgBase64),
	)
	if err != nil {
		return fmt.Errorf("LLM request: %w", err)
	}

	result := strings.TrimSpace(resp.Content)
	result = stripThinkTags(result)
	if result == "" {
		return fmt.Errorf("LLM returned empty response")
	}

	if err := s.writeClipboard(result); err != nil {
		return fmt.Errorf("write clipboard: %w", err)
	}

	questionOutputFile := "/home/shiyi/git/test/x.md"
	if err := os.MkdirAll(filepath.Dir(questionOutputFile), 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	if err := os.WriteFile(questionOutputFile, []byte(result), 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	dwmblocknotify.PUT("!!!", 5*time.Second)
	return nil
}
