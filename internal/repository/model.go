package repository

import "time"

type UserModel struct {
	id         int
	login      string
	email      string
	created_at time.Time
}
