package post

import (
	"time"

	"github.com/google/uuid"
)

// Post ...
type Post struct {
	ID       uuid.UUID
	Title    string
	Content  string
	URL      string
	Type     string
	Category string
	UserID   uuid.UUID
	Created  time.Time
	Views    int
	Votes    int
	Score    int
}
