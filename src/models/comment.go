package models

import (
	"database/sql"
	"fmt"
	"forum/utils"
	"time"
)

type Comment struct {
	ID              int
	Content         string
	UserID          int
	PostID          int
	ParentCommentID sql.NullInt64
	CreatedAt       time.Time
}

// GetCommentById récupère un commentaire par son ID
func GetCommentById(db *sql.DB, id int) (*Comment, error) {
	row := db.QueryRow("SELECT id, content, user_id, post_id, parent_comment_id, created_at FROM comment WHERE id = ?", id)

	var comment Comment
	err := row.Scan(&comment.ID, &comment.Content, &comment.UserID, &comment.PostID, &comment.ParentCommentID, &comment.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("échec de la récupération du commentaire : %v", err)
	}
	return &comment, nil
}

func CreateComment(db *sql.DB, content string, post_id int, parent_comment_id *int, user_id int) error {
	var err error

	maxId, err := utils.GetMaxID("post")
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération de l'ID maximum : %v", err)
	}

	id := maxId + 1

	if parent_comment_id != nil {
		_, err = db.Exec(
			`INSERT INTO comment (id, content, user_id, post_id, parent_comment_id, created_at) VALUES (?, ?, ?, ?, ?, NOW())`,
			id, content, user_id, post_id, *parent_comment_id)
	} else {
		_, err = db.Exec(
			`INSERT INTO comment (id, content, user_id, post_id, created_at) VALUES (?, ?, ?, ?, NOW())`,
			id, content, user_id, post_id)
	}
	if err != nil {
		return fmt.Errorf("échec de la création du commentaire: %v", err)
	}
	return nil
}
