package channel

import "context"

type MessageEvent struct {
	ChatID    string
	OpenID    string
	MsgType   string
	Content   string
	MessageID string
}

type Channel interface {
	Name() string
	SendMessage(ctx context.Context, chatID, msgType, content string) error
	StartListening(ctx context.Context) error
	StopListening()
	SetMessageHandler(handler MessageHandler)
}

type MessageHandler interface {
	OnMessageReceive(event *MessageEvent)
}
