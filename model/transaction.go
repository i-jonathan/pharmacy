package model

import (
	"encoding/json"
	"time"
)

type HeldTransaction struct {
	baseModel
	Type      string          `db:"type" json:"type"`
	Reference string          `db:"reference" json:"reference"`
	Payload   json.RawMessage `db:"payload" json:"payload"`
	UpdatedAt time.Time       `db:"updated_at" json:"updated_at"`
}
