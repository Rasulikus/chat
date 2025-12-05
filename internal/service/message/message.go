package message

import (
	"context"

	"github.com/Rasulikus/chat/internal/model"
	"github.com/Rasulikus/chat/internal/repository"
	"github.com/Rasulikus/chat/internal/service"
)

var _ service.MessageService = (*Service)(nil)

type Service struct {
	messageRepo repository.MessageRepository
}

func NewService(messageRepo repository.MessageRepository) *Service {
	return &Service{
		messageRepo: messageRepo,
	}
}

// Create creates a new message and persists it in the repository.
func (s *Service) Create(ctx context.Context, in service.CreateMessageInput) (*model.Message, error) {
	message := &model.Message{
		Nick:   in.Nick,
		Text:   in.Text,
		RoomID: in.RoomID,
	}
	err := s.messageRepo.Insert(ctx, message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

// ListByRoom returns messages for a specific room with optional pagination by beforeID and limit.
func (s *Service) ListByRoom(ctx context.Context, roomID int64, beforeID *int64, limit int) ([]model.Message, error) {
	if limit < 0 || limit > 100 {
		limit = 50
	}
	messages, err := s.messageRepo.ListByRoom(ctx, roomID, beforeID, limit)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// GetByID returns a single message by its ID.
func (s *Service) GetByID(ctx context.Context, id int64) (*model.Message, error) {
	message, err := s.messageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return message, nil
}
