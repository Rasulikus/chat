package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Message struct {
	bun.BaseModel `bun:"table:messages" swaggerignore:"true"`

	ID        int64     `json:"id" bun:"id,pk,autoincrement"`
	Nick      string    `json:"nick" bun:"nick,notnull"`
	Text      string    `json:"text" bun:"text,notnull"`
	RoomID    int64     `json:"room_id" bun:"room_id,notnull"`
	CreatedAt time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`

	Room *Room `json:"-" bun:"rel:belongs-to,join:room_id=id"`
}
