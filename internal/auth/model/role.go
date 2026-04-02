package model

import "time"

type Role struct {
	ID   int    `db:"id" json:"id"`
	Code string `db:"code" json:"code"`
	Name string `db:"name" json:"name"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	Permissions []Permission `db:"permissions" json:"permissions"`
}
