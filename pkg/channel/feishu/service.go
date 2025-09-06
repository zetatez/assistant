package feishu

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"assistant/pkg/channel"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
	"github.com/sirupsen/logrus"
)

const (
	maxSeenMessages   = 10000
	seenMessageMaxAge = 10 * time.Minute
	cleanupInterval   = 1 * time.Minute
)

type Service struct {
	client       *lark.Client
	appID        string
	appSecret    string
	handlers     []channel.MessageHandler
	seenMessages sync.Map
	logger       *logrus.Logger
	wsClient     *larkws.Client
	wsStopCh     chan struct{}
	wsStopped    bool
	wsMu         sync.Mutex
	cleanupCount int32
}

type ServiceOption func(*Service)

func WithLogger(logger *logrus.Logger) ServiceOption {
	return func(s *Service) {
		s.logger = logger
	}
}

func NewService(appID, appSecret string, opts ...ServiceOption) *Service {
	s := &Service{
		client:    lark.NewClient(appID, appSecret),
		appID:     appID,
		appSecret: appSecret,
		wsStopCh:  make(chan struct{}),
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.logger == nil {
		s.logger = logrus.New()
	}

	return s
}

func (s *Service) Name() string { return "feishu" }

func (s *Service) SendMessage(ctx context.Context, chatID, msgType, content string) error {
	createReq := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType("chat_id").
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(chatID).
			MsgType(msgType).
			Content(content).
			Build()).
		Build()

	resp, err := s.client.Im.V1.Message.Create(ctx, createReq)
	if err != nil {
		return err
	}
	if !resp.Success() {
		return err
	}
	return nil
}

func (s *Service) SetMessageHandler(handler channel.MessageHandler) {
	s.handlers = []channel.MessageHandler{handler}
}

func (s *Service) isMessageSeen(messageID string) bool {
	if _, ok := s.seenMessages.Load(messageID); ok {
		return true
	}

	if s.countSeenMessages() >= maxSeenMessages {
		s.logger.Warnf("[feishu] seenMessages reached max size %d, cleaning up", maxSeenMessages)
		s.cleanupSeenMessagesForce()
	}

	s.seenMessages.Store(messageID, time.Now())
	return false
}

func (s *Service) countSeenMessages() int {
	count := 0
	s.seenMessages.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

func (s *Service) cleanupSeenMessagesForce() {
	now := time.Now()
	deleted := 0
	s.seenMessages.Range(func(key, value any) bool {
		if now.Sub(value.(time.Time)) > seenMessageMaxAge {
			s.seenMessages.Delete(key)
			deleted++
		}
		return true
	})
	atomic.AddInt32(&s.cleanupCount, int32(deleted))
	s.logger.Infof("[feishu] force cleanup deleted %d messages, total cleanup: %d", deleted, atomic.LoadInt32(&s.cleanupCount))
}

func (s *Service) StartListening(ctx context.Context) error {
	s.logger.Infof("[feishu] starting WebSocket listener")

	eventHandler := dispatcher.NewEventDispatcher(s.appID, s.appSecret).
		OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
			if len(s.handlers) > 0 && event.Event != nil && event.Event.Message != nil {
				msg := event.Event.Message
				messageID := derefString(msg.MessageId)
				if messageID == "" || s.isMessageSeen(messageID) {
					return nil
				}
				s.dispatchEvent(&channel.MessageEvent{
					ChatID:    derefString(msg.ChatId),
					OpenID:    derefString(event.Event.Sender.SenderId.OpenId),
					MsgType:   derefString(msg.MessageType),
					Content:   derefString(msg.Content),
					MessageID: messageID,
				})
			}
			return nil
		})

	go s.cleanupSeenMessagesLoop()

	cli := larkws.NewClient(s.appID, s.appSecret,
		larkws.WithEventHandler(eventHandler),
	)

	s.wsMu.Lock()
	s.wsClient = cli
	s.wsMu.Unlock()

	return cli.Start(ctx)
}

func (s *Service) StopListening() {
	s.wsMu.Lock()
	defer s.wsMu.Unlock()

	if s.wsClient == nil || s.wsStopped {
		return
	}

	s.logger.Infof("[feishu] stopping WebSocket listener")
	s.wsStopped = true
	close(s.wsStopCh)
}

func (s *Service) cleanupSeenMessagesLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			deleted := 0
			s.seenMessages.Range(func(key, value any) bool {
				if now.Sub(value.(time.Time)) > seenMessageMaxAge {
					s.seenMessages.Delete(key)
					deleted++
				}
				return true
			})
			if deleted > 0 {
				atomic.AddInt32(&s.cleanupCount, int32(deleted))
				s.logger.Debugf("[feishu] cleanup deleted %d old messages, total: %d", deleted, atomic.LoadInt32(&s.cleanupCount))
			}
		case <-s.wsStopCh:
			s.logger.Infof("[feishu] seenMessages cleanup stopped, total cleaned: %d", atomic.LoadInt32(&s.cleanupCount))
			return
		}
	}
}

func (s *Service) dispatchEvent(event *channel.MessageEvent) {
	for _, h := range s.handlers {
		h.OnMessageReceive(event)
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
