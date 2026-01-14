package llm

import (
	"context"
	"fmt"
	"log"
	"os"
)

// func main() {
// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()
//
// 	client := qwen.New(qwen.Config{
// 		APIKey: os.Getenv("QWEN_API_KEY"),
// 		Model:  "qwen-turbo",
// 	})
//
// 	req := llm.ChatRequest{
// 		Messages: []llm.Message{
// 			{Role: "system", Content: "You are a helpful assistant"},
// 			{Role: "user", Content: "用一句话解释什么是 Go 接口"},
// 		},
// 	}
//
// 	resp, err := llm.Chat(ctx, client, req, nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	fmt.Println(resp.Choices[0].Message.Content)
// }

func main() {
	ctx := context.Background()

	client := openai.New(openai.Config{
		APIKey: os.Getenv("OPENAI_API_KEY"),
		Model:  "gpt-4o-mini",
	})

	req := llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "user", Content: "逐步推理：如何实现一个 Go LRU 缓存？"},
		},
	}

	_, err := llm.Chat(
		ctx,
		client,
		req,
		func(delta llm.StreamDelta) {
			fmt.Print(delta.Content)
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}
