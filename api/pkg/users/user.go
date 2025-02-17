package users

import "time"

type User struct {
	ID            string     `db:"id" json:"id"`
	Name          string     `db:"name" json:"name"`
	Email         string     `db:"email" json:"email"`
	EmailVerified bool       `db:"email_verified" json:"email_verified"`
	PasswordHash  string     `db:"password" json:"-"`
	CreatedAt     *time.Time `db:"created_at" json:"created_at"`
	AvatarURL     *string    `db:"avatar_url" json:"avatar_url"`
}
