package user

import "github.com/google/uuid"

// User ...
type User struct {
	ID       uuid.UUID
	Username string
	Password string
}
