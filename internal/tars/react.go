package tars

import (
	"context"
	"fmt"

	llm "assistant/pkg/llmproxy"
)

const maxReActIterations = 5

func runReAct(ctx context.Context, client llm.Client, req llm.ChatRequest, executor *ToolExecutor, logger Logger) (string, error) {
	messages := req.Messages

	for i := 0; i < maxReActIterations; i++ {
		req.Messages = messages
		resp, err := client.Chat(ctx, req)
		if err != nil {
			return "", err
		}

		if len(resp.ToolCalls) == 0 {
			return resp.Content, nil
		}

		messages = append(messages, llm.Message{
			Role:      llm.RoleAI,
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		for _, tc := range resp.ToolCalls {
			result, err := executor.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
			if err != nil {
				logger.Warnf("tars: tool %s error: %v", tc.Function.Name, err)
				result = fmt.Sprintf("Tool error: %v", err)
			}
			messages = append(messages, llm.Message{
				Role:       llm.RoleTool,
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}

	return "", fmt.Errorf("max iterations (%d) reached without final answer", maxReActIterations)
}
