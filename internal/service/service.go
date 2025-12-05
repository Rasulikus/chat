package service

import (
	"context"
	"time"

	"github.com/Rasulikus/chat/internal/model"
)

type CreateRoomInput struct {
	Name     string
	Password string
}

type UpdateRoomInput struct {
	ID       int64
	Name     *string
	Password *string
}

type RoomService interface {
	Create(ctx context.Context, in CreateRoomInput) (*model.Room, error)
	GetByID(ctx context.Context, id int64) (*model.Room, error)
	List(ctx context.Context, limit int, order string, beforeID *int64) ([]model.Room, error)
	TouchActivity(ctx context.Context, id int64) error
	SoftDeleteInactiveOlderThan(ctx context.Context, olderThan time.Duration) (int64, error)
	SoftDelete(ctx context.Context, id int64) error
	CheckPassword(ctx context.Context, id int64, password string) (bool, error)
}

type CreateMessageInput struct {
	RoomID int64
	Nick   string
	Text   string
}

type MessageService interface {
	Create(ctx context.Context, in CreateMessageInput) (*model.Message, error)
	GetByID(ctx context.Context, id int64) (*model.Message, error)
	ListByRoom(ctx context.Context, roomID int64, beforeID *int64, limit int) ([]model.Message, error)
}
