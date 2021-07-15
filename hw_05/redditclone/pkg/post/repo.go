package post

import "github.com/google/uuid"

// PostRepo ...
type PostRepo interface {
	Post(id uuid.UUID) (Post, error)
	Posts() ([]Post, error)
	PostsByCategory(cat string) ([]Post, error)
	PostsByAuthor(userID uuid.UUID) ([]Post, error)
	Create(p *Post) error
	Update(p *Post) error
	Delete(id uuid.UUID) error
}
