package models

import (
	"database/sql"
	"fmt"
	"forum/utils"
	"time"
)

type Vote struct {
	ID        int
	UserID    int
	PostID    sql.NullInt64
	CommentID sql.NullInt64
	Value     int
	CreatedAt time.Time
}

// GetVoteById récupère un vote par son ID
func GetVoteById(db *sql.DB, id int) (*Vote, error) {
	row := db.QueryRow("SELECT id, user_id, post_id, comment_id, value, created_at FROM vote WHERE id = ?", id)

	var vote Vote
	err := row.Scan(&vote.ID, &vote.UserID, &vote.PostID, &vote.CommentID, &vote.Value, &vote.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("échec de la récupération du vote : %v", err)
	}
	return &vote, nil
}

func CreateVote(db *sql.DB, user_id int, post_id *int, comment_id *int, value int) error {
	maxId, err := utils.GetMaxID("post")
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération de l'ID maximum : %v", err)
	}

	id := maxId + 1

	if post_id != nil {
		_, err = db.Exec(
			`INSERT INTO vote (id, user_id, post_id, value, created_at) VALUES (?, ?, ?, ?, NOW())`,
			id, user_id, *post_id, value)
	} else if comment_id != nil {
		_, err = db.Exec(
			`INSERT INTO vote (id, user_id, comment_id, value, created_at) VALUES (?, ?, ?, ?, NOW())`,
			id, user_id, *comment_id, value)
	} else {
		return fmt.Errorf("au moins un des paramètres post_id ou comment_id doit être non-nul")
	}
	if err != nil {
		return fmt.Errorf("échec de la création du vote: %v", err)
	}
	return nil
}
