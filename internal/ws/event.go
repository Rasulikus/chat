package ws

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Rasulikus/chat/internal/model"
)

const (
	EventTypeJoin    = "join"
	EventTypeMessage = "message"
	EventTypeHistory = "load_history"
	EventTypeError   = "error"
)

type IncomingEvent struct {
	Type     string `json:"type"`
	RoomID   int64  `json:"room_id,omitempty"`
	Nick     string `json:"nick,omitempty"`
	Text     string `json:"text,omitempty"`
	BeforeID *int64 `json:"before_id,omitempty"`
	Password string `json:"password,omitempty"`
}

type OutgoingEvent struct {
	Type     string          `json:"type"`
	RoomID   int64           `json:"room_id,omitempty"`
	Message  *model.Message  `json:"message,omitempty"`
	Messages []model.Message `json:"messages,omitempty"`
	Nick     string          `json:"nick,omitempty"`
	Text     string          `json:"text,omitempty"`
}

var (
	ErrUnknownType = errors.New("unknown event type")
	ErrBadPayload  = errors.New("invalid event payload")
)

func (e *IncomingEvent) Validate() error {
	switch e.Type {
	case EventTypeJoin:
		return e.validateJoin()
	case EventTypeMessage:
		return e.validateMessage()
	case EventTypeHistory:
		return e.validateLoadHistory()
	default:
		return ErrUnknownType
	}
}

func (e *IncomingEvent) validateJoin() error {
	if e.RoomID == 0 {
		return fmt.Errorf("%w: room_id is required for join", ErrBadPayload)
	}
	if strings.TrimSpace(e.Nick) == "" {
		return fmt.Errorf("%w: nick is required for join", ErrBadPayload)
	}
	return nil
}

func (e *IncomingEvent) validateMessage() error {
	if strings.TrimSpace(e.Text) == "" {
		return fmt.Errorf("%w: text is required for message", ErrBadPayload)
	}
	return nil
}

func (e *IncomingEvent) validateLoadHistory() error {
	return nil
}
