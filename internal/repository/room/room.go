package room

import (
	"context"
	"time"

	"github.com/Rasulikus/chat/internal/model"
	"github.com/Rasulikus/chat/internal/repository"
	"github.com/uptrace/bun"
)

var _ repository.RoomRepository = (*Repository)(nil)

type Repository struct {
	db *bun.DB
}

func NewRepository(db *bun.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Insert(ctx context.Context, room *model.Room) error {
	_, err := r.db.NewInsert().Model(room).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*model.Room, error) {
	room := new(model.Room)

	err := r.db.NewSelect().Model(room).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, repository.IsNoRowsError(err)
	}
	return room, nil
}

func (r *Repository) List(ctx context.Context, limit int, order string, beforeID *int64) ([]model.Room, error) {
	var rooms []model.Room
	q := r.db.NewSelect().
		Model(&rooms)

	if beforeID != nil {
		q.Where("id < ?", *beforeID)
	}

	err := q.
		Order(order).
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return rooms, nil
}

// TouchActivity updates the activity timestamp of a room by its ID.
func (r *Repository) TouchActivity(ctx context.Context, id int64) error {
	res, err := r.db.NewUpdate().
		Model((*model.Room)(nil)).
		Set("updated_at = current_timestamp").
		Set("last_active_at = current_timestamp").
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return model.ErrNotFound
	}

	return nil
}

// SoftDeleteInactiveOlderThan soft deletes rooms that have been inactive longer than d.
// It returns the number of rooms that were marked as deleted.
func (r *Repository) SoftDeleteInactiveOlderThan(ctx context.Context, d time.Duration) (int64, error) {
	res, err := r.db.NewDelete().
		Model((*model.Room)(nil)).
		Where("last_active_at < ?", time.Now().Add(-d)).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *Repository) SoftDelete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*model.Room)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}
