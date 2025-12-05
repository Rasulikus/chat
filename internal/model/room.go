package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Room struct {
	bun.BaseModel `bun:"table:rooms" swaggerignore:"true"`

	ID           int64  `json:"id" bun:"id,pk,autoincrement"`
	Name         string `json:"name" bun:"name,notnull"`
	PasswordHash []byte `json:"-" bun:"password_hash,nullzero"`
	HasPassword  bool   `json:"has_password" bun:"-"`

	CreatedAt time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp"`
	DeletedAt time.Time `json:"deleted_at" bun:"deleted_at,soft_delete,nullzero"`

	LastActiveAt time.Time `json:"last_active_at" bun:"last_active_at,notnull,default:current_timestamp"`
}
