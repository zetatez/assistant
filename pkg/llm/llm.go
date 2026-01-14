package llm

import (
	"context"
)

// Chat is a convenience function that handles both streaming and non‑streaming requests.
// If a callback is provided and the client supports streaming, it will use StreamChat.
// Otherwise, it falls back to a regular Chat call.
func Chat(ctx context.Context, client Client, req ChatRequest, cb StreamCallback) (*ChatResponse, error) {
	// If no callback is requested, always use regular chat
	if cb == nil {
		return client.Chat(ctx, req)
	}

	// Check if the client supports streaming
	caps := client.Capabilities()
	if !caps.Has(CapabilityStream) {
		// Fall back to regular chat, but ignore the callback
		return client.Chat(ctx, req)
	}

	// Try streaming; if it returns ErrNotImplemented, fall back
	err := client.StreamChat(ctx, req, cb)
	if err == ErrNotImplemented {
		return client.Chat(ctx, req)
	}
	if err != nil {
		return nil, err
	}

	// Streaming succeeded, return nil to indicate no complete response
	return nil, nil
}

// ErrNotImplemented is returned when a capability is not implemented by a client.
var ErrNotImplemented = &notImplementedError{}

type notImplementedError struct{}

func (e *notImplementedError) Error() string {
	return "not implemented"
}
