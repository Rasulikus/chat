package message

import (
	"context"

	"github.com/Rasulikus/chat/internal/model"
	"github.com/Rasulikus/chat/internal/repository"
	"github.com/uptrace/bun"
)

var _ repository.MessageRepository = (*Repository)(nil)

type Repository struct {
	db *bun.DB
}

func NewRepository(db *bun.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Insert(ctx context.Context, message *model.Message) error {
	_, err := r.db.NewInsert().Model(message).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*model.Message, error) {
	message := new(model.Message)
	err := r.db.NewSelect().Model(message).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, repository.IsNoRowsError(err)
	}
	return message, nil
}

func (r *Repository) ListByRoom(ctx context.Context, roomID int64, beforeID *int64, limit int) ([]model.Message, error) {
	var messages []model.Message
	q := r.db.NewSelect().
		Model(&messages).
		Where("room_id = ?", roomID)

	if beforeID != nil {
		q.Where("id < ?", *beforeID)
	}

	err := q.
		Order("id ASC").
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return messages, nil
}
