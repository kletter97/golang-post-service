package repository

import ("context"
		"time"
)

type Post struct {
	ID        int64  `json:"id"`
	AuthorID  int64  `json:"author_id"`
	Content   string `json:"content"`
	Status    string `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type PostRepository interface {
	Create(ctx context.Context, author_id int64, content string) (*Post, error)
	GetPostsByAuthor(ctx context.Context, author_id int64) ([]Post, error)
}
