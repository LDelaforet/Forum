package utils

import (
	"database/sql"
	"time"
)

type User struct {
	ID         int
	Username   string
	Email      string
	PasswdHash string
	Salt       string
	CreatedAt  time.Time
}

type Post struct {
	ID        int
	Title     string
	Content   string
	UserID    int
	CreatedAt time.Time
}

type Comment struct {
	ID              int
	Content         string
	UserID          int
	PostID          sql.NullInt64
	ParentCommentID sql.NullInt64
	CreatedAt       time.Time
}

type Vote struct {
	ID        int
	UserID    int
	PostID    sql.NullInt64
	CommentID sql.NullInt64
	Value     int
	CreatedAt time.Time
}
