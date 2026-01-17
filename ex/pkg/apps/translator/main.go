package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"assistant/pkg/apps"
	"assistant/pkg/llm"

	_ "assistant/pkg/llm/providers/deepseek"
)

func main() {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set the DEEPSEEK_API_KEY environment variable")
	}

	cfg := llm.Config{
		APIKey:  apiKey,
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-chat",
	}

	client, err := llm.NewClient("deepseek", cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	tr := apps.NewTranslator(client)

	ctx := context.Background()

	translateAndPrint(ctx, tr, "quantum", "zh")

	translateAndPrint(
		ctx,
		tr,
		"量子计算机利用量子比特的叠加与纠缠特性处理信息。",
		"en",
	)

	translateAndPrint(
		ctx,
		tr,
		`量子计算是一种基于量子力学原理的计算模型。
它利用叠加、纠缠等特性，在特定问题上可以显著提升计算效率。`,
		"en",
	)
}

func translateAndPrint(
	ctx context.Context,
	tr *apps.Translator,
	text string,
	targetLang string,
) {
	result, err := tr.Translate(ctx, text, targetLang)
	if err != nil {
		log.Printf("Translation failed: %v\n", err)
		return
	}

	printJSONAndMarkdown(result)
}

func printJSONAndMarkdown(v interface{}) {
	// JSON 输出
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(v)

	// Markdown 输出
	os.Stdout.WriteString("\n## 翻译结果\n\n")

	if r, ok := v.(*apps.TranslateResult); ok {
		os.Stdout.WriteString(fmt.Sprintf("- **源语言**: %s\n", r.SourceLanguage))
		os.Stdout.WriteString(fmt.Sprintf("- **目标语言**: %s\n", r.TargetLanguage))
		os.Stdout.WriteString(fmt.Sprintf("- **输入类型**: %s\n", r.InputType))
		os.Stdout.WriteString(fmt.Sprintf("- **翻译**: %s\n", r.Translation))
		os.Stdout.WriteString(fmt.Sprintf("- **置信度**: %.2f\n\n", r.Confidence))
	}

	os.Stdout.WriteString("\n---\n\n")
}
