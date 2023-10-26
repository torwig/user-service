package entities

import "time"

type User struct {
	ID          int64
	FirstName   string
	LastName    string
	PhoneNumber string
	Address     string
	Deleted     bool
	CreatedAt   time.Time
	DeletedAt   *time.Time
}

func (u User) IsDeleted() bool {
	return u.Deleted
}

type CreateUserParams struct {
	FirstName   string
	LastName    string
	PhoneNumber string
	Address     string
}

type UpdateUserParams struct {
	FirstName   *string
	LastName    *string
	PhoneNumber *string
	Address     *string
}
