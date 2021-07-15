package comment

import "github.com/google/uuid"

// CommentRepo ...
type CommentRepo interface {
	Comment(id uuid.UUID) Comment
	CommentsByPost(postID uuid.UUID) ([]Comment, error)
	Create(c *Comment) error
	Update(c *Comment) error
	Delete(id uuid.UUID) error
}
