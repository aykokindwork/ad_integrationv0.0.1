package model

import "time"

type User struct {
	ID    int    `db:"id"`
	Login string `db:"login"`
	Email string `db:"email"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	RoleList []Role `db:"roles"`
}
