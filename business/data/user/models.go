package user

import (
	"time"
)

// User represents an individual user.
type User struct {
	ID           string    `db:"user_id" json:"id"`
	Name         string    `db:"name" json:"name"`
	Email        string    `db:"email" json:"email"`
	Roles        []string  `db:"roles" json:"roles"`
	PasswordHash []byte    `db:"password_hash" json:"-"`
	DateCreated  time.Time `db:"date_created" json:"date_created"`
	DateUpdated  time.Time `db:"date_updated" json:"date_updated"`
}

type NewUser struct {
	Name            string   `json:"name" validate:"required"`
	Email           string   `json:"email" validate:"required,email"`
	Roles           []string `json:"roles" validate:"required"`
	Password        string   `json:"password" validate:"required"`
	PasswordConfirm string   `json:"password_confirm" validate:"eqfield=Password"`
}
