package user

import "github.com/google/uuid"

// UserRepo ...
type UserRepo interface {
	User(id uuid.UUID) (User, error)
	UserByUsername(un string) (User, error)
	Create(u *User) error
	Update(u *User) error
	Delete(id uuid.UUID) error
}
