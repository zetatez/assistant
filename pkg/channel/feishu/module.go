package feishu

import (
	"context"

	"assistant/internal/bootstrap/psl"
	"assistant/pkg/channel"

	"github.com/gin-gonic/gin"
)

type Module struct {
	service *Service
}

func NewModule(appID, appSecret string) *Module {
	return &Module{
		service: NewService(appID, appSecret, WithLogger(psl.GetLogger())),
	}
}

func (m *Module) Name() string {
	return m.service.Name()
}

func (m *Module) SendMessage(ctx context.Context, chatID, msgType, content string) error {
	return m.service.SendMessage(ctx, chatID, msgType, content)
}

func (m *Module) StartListening(ctx context.Context) error {
	return m.service.StartListening(ctx)
}

func (m *Module) StopListening() {
	m.service.StopListening()
}

func (m *Module) SetMessageHandler(handler channel.MessageHandler) {
	m.service.SetMessageHandler(handler)
}

func (m *Module) Register(r *gin.RouterGroup) {
}

func (m *Module) Middleware() []gin.HandlerFunc {
	return nil
}

func (m *Module) AsChannel() channel.Channel {
	return m.service
}
