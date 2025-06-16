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

// CreateVote crée un nouveau vote
func (v *Vote) CreateVote() error {
	query := "INSERT INTO vote (user_id, post_id, value) VALUES (?, ?, ?)"
	result, err := DbContext.Exec(query, v.UserID, v.PostID, v.Value)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du vote : %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération de l'ID : %v", err)
	}
	v.ID = int(id)
	return nil
}

// GetVoteByUserAndPost récupère le vote d'un utilisateur sur un post
func GetVoteByUserAndPost(userID, postID int) (*Vote, error) {
	vote := &Vote{}
	query := "SELECT id, user_id, post_id, value FROM vote WHERE user_id = ? AND post_id = ?"
	err := DbContext.QueryRow(query, userID, postID).Scan(&vote.ID, &vote.UserID, &vote.PostID, &vote.Value)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération du vote : %v", err)
	}
	return vote, nil
}

// UpdateVote met à jour un vote existant
func (v *Vote) UpdateVote() error {
	query := "UPDATE vote SET value = ? WHERE id = ?"
	_, err := DbContext.Exec(query, v.Value, v.ID)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour du vote : %v", err)
	}
	return nil
}

// DeleteVote supprime un vote
func (v *Vote) DeleteVote() error {
	query := "DELETE FROM vote WHERE id = ?"
	_, err := DbContext.Exec(query, v.ID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression du vote : %v", err)
	}
	return nil
}

// GetPostVotes récupère tous les votes d'un post
func GetPostVotes(postID int) ([]*Vote, error) {
	query := "SELECT id, user_id, post_id, value FROM vote WHERE post_id = ?"
	rows, err := DbContext.Query(query, postID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des votes : %v", err)
	}
	defer rows.Close()

	var votes []*Vote
	for rows.Next() {
		vote := &Vote{}
		err := rows.Scan(&vote.ID, &vote.UserID, &vote.PostID, &vote.Value)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture des votes : %v", err)
		}
		votes = append(votes, vote)
	}
	return votes, nil
}

// GetPostScore calcule le score total d'un post
func GetPostScore(postID int) (int, error) {
	query := "SELECT COALESCE(SUM(value), 0) FROM vote WHERE post_id = ?"
	var score int
	err := DbContext.QueryRow(query, postID).Scan(&score)
	if err != nil {
		return 0, fmt.Errorf("erreur lors du calcul du score du post : %v", err)
	}
	return score, nil
}

// GetUserVotes récupère tous les votes d'un utilisateur
func GetUserVotes(userID int) ([]*Vote, error) {
	query := "SELECT id, user_id, post_id, value FROM vote WHERE user_id = ?"
	rows, err := DbContext.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des votes : %v", err)
	}
	defer rows.Close()

	var votes []*Vote
	for rows.Next() {
		vote := &Vote{}
		err := rows.Scan(&vote.ID, &vote.UserID, &vote.PostID, &vote.Value)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture des votes : %v", err)
		}
		votes = append(votes, vote)
	}
	return votes, nil
}
