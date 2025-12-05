package repository

import (
	"context"
	"time"

	"github.com/Rasulikus/chat/internal/model"
)

type RoomRepository interface {
	Insert(ctx context.Context, room *model.Room) error
	GetByID(ctx context.Context, id int64) (*model.Room, error)
	List(ctx context.Context, limit int, order string, beforeID *int64) ([]model.Room, error)
	TouchActivity(ctx context.Context, roomID int64) error
	SoftDeleteInactiveOlderThan(ctx context.Context, d time.Duration) (int64, error)
	SoftDelete(ctx context.Context, id int64) error
}

type MessageRepository interface {
	Insert(ctx context.Context, message *model.Message) error
	GetByID(ctx context.Context, id int64) (*model.Message, error)
	ListByRoom(ctx context.Context, roomID int64, beforeID *int64, limit int) ([]model.Message, error)
}
