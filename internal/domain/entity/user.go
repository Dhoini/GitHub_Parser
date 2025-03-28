package entity

import "time"

type User struct {
	ID        int64
	Login     string
	Name      string
	Email     string
	AvatarURL string
	Bio       string
	Company   string
	Location  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
