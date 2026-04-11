package channel

import "context"

type MessageEvent struct {
	SessionID string
	OpenID    string
	MsgType   string
	Content   string
	MessageID string
}

type Channel interface {
	Name() string
	SendMessage(ctx context.Context, sessionID, msgType, content string) error
	DownloadMedia(ctx context.Context, messageID, fileKey string) ([]byte, string, error)
	StartListening(ctx context.Context) error
	StopListening()
	SetMessageHandler(handler MessageHandler)
}

type MessageHandler interface {
	OnMessageReceive(event *MessageEvent)
}
