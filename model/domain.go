package model

import "time"

type baseModel struct {
	ID        int
	CreatedAt time.Time
}

type passwordConfig struct {
	time       uint32
	memory     uint32
	threads    uint8
	keyLength  uint32
	saltLength uint32
}
