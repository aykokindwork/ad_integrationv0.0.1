package model

import "time"

type User struct {
	ID    int    `db:"id" json:"id"`
	Login string `db:"login" json:"login"`
	Email string `db:"email" json:"email"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	Roles []Role `db:"roles" json:"roles"`
}
