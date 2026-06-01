package repository

import "context"

type Post struct {
	ID        int64  `json:"id"`
	AuthorID  int64  `json:"author_id"`
	Content   string `json:"content"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type PostRepository interface {
	Create(ctx context.Context, author_id int64, content string) (*Post, error)
	//GetByAuthor(ctx context.Context, email string) (*Post, error) //??
}
