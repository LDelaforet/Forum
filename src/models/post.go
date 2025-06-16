package models

import (
	"time"
)

type Post struct {
	ID        int
	Title     string
	Content   string
	UserID    int
	CreatedAt time.Time

	// Seront definis plus tard, je sais que c'est atroce mais j'ai pas trop le choix
	VoteScore int
	Username  string
}
