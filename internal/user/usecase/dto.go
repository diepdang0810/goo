package usecase

import "time"

type CreateUserInput struct {
	Name  string
	Email string
}

type UserOutput struct {
	ID        int64
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
