package comment

import (
	"time"

	"github.com/google/uuid"
)

// Comment ...
type Comment struct {
	ID      uuid.UUID
	PostID  uuid.UUID
	Content string
	UserID  uuid.UUID
	Created time.Time
}
