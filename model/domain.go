package model

import "time"

type baseModel struct {
	ID        int       `db:"id"`
	CreatedAt time.Time `db:"created_at"`
}

type passwordConfig struct {
	time       uint32
	memory     uint32
	keyLength  uint32
	saltLength uint32
	threads    uint8
}
