package llm

import (
	"context"
	"errors"
)

func Chat(
	ctx context.Context,
	c Client,
	req ChatRequest,
	cb StreamCallback,
) (*ChatResponse, error) {
	caps := c.Capabilities()

	// 1️⃣ 如果提供了回调并且客户端支持流式，尝试流式聊天
	if cb != nil && caps.Has(CapabilityStream) {
		err := c.StreamChat(ctx, req, cb)
		if err == nil {
			return nil, nil
		}
		// 如果返回 ErrNotImplemented，回退到普通 Chat
		if errors.Is(err, ErrNotImplemented) {
			// 继续执行普通 Chat
		} else {
			return nil, err
		}
	}

	// 2️⃣ 回退到普通 Chat
	if !caps.Has(CapabilityChat) {
		return nil, ErrNotImplemented
	}

	return c.Chat(ctx, req)
}
