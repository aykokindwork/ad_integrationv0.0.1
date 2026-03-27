package model

import "time"

type Permission struct {
	ID   int    `db:"id"`
	Code string `db:"code"`
	Name string `db:"name"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
