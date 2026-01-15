package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"assistant/pkg/llm"
	"assistant/pkg/llmtranslator"

	_ "assistant/pkg/llm/providers/deepseek"
)

func main() {
	// 1. Read API key
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set the DEEPSEEK_API_KEY environment variable")
	}

	// 2. Create LLM config
	cfg := llm.Config{
		APIKey:  apiKey,
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-chat",
	}

	// 3. Create LLM client
	client, err := llm.NewClient("deepseek", cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Printf("Using provider: %s\n", client.Provider())
	fmt.Printf("Using model: %s\n", client.Model())

	// 4. Create translator
	tr := llmtranslator.New(client)

	ctx := context.Background()

	// ---------- Example 1: Word translation ----------
	fmt.Println("\n--- Word Translation ---")
	translateAndPrint(ctx, tr, "quantum", "zh")

	// ---------- Example 2: Sentence translation ----------
	fmt.Println("\n--- Sentence Translation ---")
	translateAndPrint(
		ctx,
		tr,
		"量子计算机利用量子比特的叠加与纠缠特性处理信息。",
		"en",
	)

	// ---------- Example 3: Article translation ----------
	fmt.Println("\n--- Article Translation ---")
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
	tr *llmtranslator.Translator,
	text string,
	targetLang string,
) {
	result, err := tr.Translate(ctx, text, targetLang)
	if err != nil {
		log.Printf("Translation failed: %v\n", err)
		return
	}

	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(b))
}
