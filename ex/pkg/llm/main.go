// This is a simple example demonstrating how to use the DeepSeek LLM provider.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"assistant/pkg/llm"

	_ "assistant/pkg/llm/providers/deepseek"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set the DEEPSEEK_API_KEY environment variable")
	}

	// Create configuration for DeepSeek
	cfg := llm.Config{
		APIKey:  apiKey,
		BaseURL: "https://api.deepseek.com", // Optional, defaults to this
		Model:   "deepseek-chat",            // Optional, defaults to this
	}

	// Create a DeepSeek client
	client, err := llm.NewClient("deepseek", cfg)
	if err != nil {
		log.Fatalf("Failed to create DeepSeek client: %v", err)
	}

	fmt.Printf("Using provider: %s\n", client.Provider())
	fmt.Printf("Using model: %s\n", client.Model())

	// Prepare a chat request
	req := llm.ChatRequest{
		Model: cfg.Model,
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "帮我介绍量子计算机原理，使用一句话"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	// Make the request using the client's Chat method
	ctx := context.Background()

	// First, try regular chat
	fmt.Println("\n--- Regular Chat ---")
	resp, err := client.Chat(ctx, req)
	if err != nil {
		log.Printf("Chat request failed: %v", err)
	} else {
		fmt.Println("Response:")
		fmt.Println(resp.Content)
		if resp.Usage.TotalTokens > 0 {
			fmt.Printf("Token usage: Prompt=%d, Completion=%d, Total=%d\n",
				resp.Usage.PromptTokens,
				resp.Usage.CompletionTokens,
				resp.Usage.TotalTokens)
		}
	}

	// Try streaming if supported
	caps := client.Capabilities()
	if caps.Has(llm.CapabilityStream) {
		fmt.Println("\n--- Streaming Chat ---")
		streamReq := llm.ChatRequest{
			Model: cfg.Model,
			Messages: []llm.Message{
				{Role: llm.RoleUser, Content: "帮我介绍量子计算机原理, 使用三句话"},
			},
			Temperature: 0.7,
			MaxTokens:   20,
		}

		fmt.Print("Streaming response: ")
		err := client.StreamChat(ctx, streamReq, func(chunk llm.ChatResponse) {
			fmt.Print(chunk.Content)
		})
		if err != nil {
			log.Printf("Streaming chat failed: %v", err)
		}
		fmt.Println()
	} else {
		fmt.Println("\nNote: This provider doesn't support streaming")
	}
}
