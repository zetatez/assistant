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

	// 1️⃣ 优先走 Stream
	if cb != nil && caps.Has(CapabilityStream) {
		if err := c.StreamChat(ctx, req, cb); err == nil {
			return nil, nil
		} else if !errors.Is(err, ErrNotImplemented) {
			return nil, err
		}
	}

	// 2️⃣ fallback 到普通 Chat
	if !caps.Has(CapabilityChat) {
		return nil, ErrNotImplemented
	}

	return c.Chat(ctx, req)
}
