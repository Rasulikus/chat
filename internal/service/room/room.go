package room

import (
	"context"
	"errors"
	"time"

	"github.com/Rasulikus/chat/internal/model"
	"github.com/Rasulikus/chat/internal/repository"
	"github.com/Rasulikus/chat/internal/service"
	"golang.org/x/crypto/bcrypt"
)

var _ service.RoomService = (*Service)(nil)

type Service struct {
	roomRepo repository.RoomRepository
}

func NewService(roomRepo repository.RoomRepository) *Service {
	return &Service{
		roomRepo: roomRepo,
	}
}

// Create creates a new room, hashes the password if provided, and persists it in the repository.
func (s *Service) Create(ctx context.Context, in service.CreateRoomInput) (*model.Room, error) {
	var hashedPassword []byte
	var err error
	hasPassword := false
	if in.Password != "" {
		hashedPassword, err = bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		hasPassword = true
	}

	room := &model.Room{
		Name:         in.Name,
		PasswordHash: hashedPassword,
		HasPassword:  hasPassword,
	}

	err = s.roomRepo.Insert(ctx, room)
	if err != nil {
		return nil, err
	}
	return room, nil
}

// List returns a list of rooms with optional pagination and ordering, and marks which rooms are password protected.
func (s *Service) List(ctx context.Context, limit int, order string, beforeID *int64) ([]model.Room, error) {
	if limit == 0 {
		limit = 20
	}
	rooms, err := s.roomRepo.List(ctx, limit, order, beforeID)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(rooms); i++ {
		if rooms[i].PasswordHash != nil {
			rooms[i].HasPassword = true
		}
	}

	return rooms, nil
}

// GetByID returns a room by its ID and indicates whether it is password protected.
func (s *Service) GetByID(ctx context.Context, id int64) (*model.Room, error) {
	room, err := s.roomRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if room.PasswordHash != nil {
		room.HasPassword = true
	}

	return room, nil
}

// TouchActivity updates the activity timestamp of a room by its ID.
func (s *Service) TouchActivity(ctx context.Context, roomID int64) error {
	return s.roomRepo.TouchActivity(ctx, roomID)
}

// SoftDeleteInactiveOlderThan soft deletes rooms that have been inactive longer than d.
// It returns the number of rooms that were marked as deleted.
func (s *Service) SoftDeleteInactiveOlderThan(ctx context.Context, olderThan time.Duration) (int64, error) {
	return s.roomRepo.SoftDeleteInactiveOlderThan(ctx, olderThan)
}

// SoftDelete performs a soft delete of a room by its ID.
func (s *Service) SoftDelete(ctx context.Context, id int64) error {
	err := s.roomRepo.SoftDelete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

// CheckPassword verifies a plain-text password against the stored room password hash and returns whether they match.
func (s *Service) CheckPassword(ctx context.Context, id int64, password string) (bool, error) {
	room, err := s.roomRepo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	if room.PasswordHash == nil {
		return true, nil
	}

	err = bcrypt.CompareHashAndPassword(room.PasswordHash, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
