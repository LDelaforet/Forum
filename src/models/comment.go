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
	Username        string // Pour l'affichage
}

// GetCommentById récupère un commentaire par son ID
func GetCommentById(db *sql.DB, id int) (*Comment, error) {
	query := `
		SELECT c.id, c.content, c.user_id, c.post_id, c.parent_comment_id, c.created_at, u.username 
		FROM comment c 
		JOIN user u ON c.user_id = u.id 
		WHERE c.id = ?`

	comment := &Comment{}
	var parentID sql.NullInt64
	err := db.QueryRow(query, id).Scan(
		&comment.ID, &comment.Content, &comment.UserID, &comment.PostID,
		&parentID, &comment.CreatedAt, &comment.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("échec de la récupération du commentaire : %v", err)
	}
	comment.ParentCommentID = parentID
	return comment, nil
}

// CreateComment crée un nouveau commentaire
func CreateComment(db *sql.DB, content string, post_id int, parent_comment_id *int, user_id int) error {
	var err error

	maxId, err := utils.GetMaxID("comment")
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

// GetCommentsByPostID récupère tous les commentaires d'un post
func GetCommentsByPostID(postID int) ([]*Comment, error) {
	query := `
		SELECT c.id, c.content, c.user_id, c.post_id, c.parent_comment_id, c.created_at, u.username 
		FROM comment c 
		JOIN user u ON c.user_id = u.id 
		WHERE c.post_id = ? 
		ORDER BY c.created_at ASC`

	rows, err := DbContext.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		comment := &Comment{}
		var parentID sql.NullInt64
		err := rows.Scan(&comment.ID, &comment.Content, &comment.UserID, &comment.PostID, &parentID, &comment.CreatedAt, &comment.Username)
		if err != nil {
			return nil, err
		}
		comment.ParentCommentID = parentID
		comments = append(comments, comment)
	}
	return comments, nil
}

// GetCommentByID récupère un commentaire par son ID
func GetCommentByID(commentID int) (*Comment, error) {
	query := `
		SELECT c.id, c.content, c.user_id, c.post_id, c.parent_comment_id, c.created_at, u.username 
		FROM comment c 
		JOIN user u ON c.user_id = u.id 
		WHERE c.id = ?`

	comment := &Comment{}
	var parentID sql.NullInt64
	err := DbContext.QueryRow(query, commentID).Scan(
		&comment.ID, &comment.Content, &comment.UserID, &comment.PostID,
		&parentID, &comment.CreatedAt, &comment.Username)
	if err != nil {
		return nil, err
	}
	comment.ParentCommentID = parentID
	return comment, nil
}

// UpdateComment met à jour un commentaire
func (c *Comment) UpdateComment() error {
	query := "UPDATE comment SET content = ? WHERE id = ?"
	_, err := DbContext.Exec(query, c.Content, c.ID)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour du commentaire : %v", err)
	}
	return nil
}

// DeleteComment supprime un commentaire
func (c *Comment) DeleteComment() error {
	query := "DELETE FROM comment WHERE id = ?"
	_, err := DbContext.Exec(query, c.ID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression du commentaire : %v", err)
	}
	return nil
}

// GetPostComments récupère tous les commentaires d'un post
func GetPostComments(postID int) ([]*Comment, error) {
	query := "SELECT id, content, user_id, post_id, created_at FROM comments WHERE post_id = ? ORDER BY created_at DESC"
	rows, err := DbContext.Query(query, postID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des commentaires : %v", err)
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		comment := &Comment{}
		err := rows.Scan(&comment.ID, &comment.Content, &comment.UserID, &comment.PostID, &comment.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture des commentaires : %v", err)
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

// GetUserComments récupère tous les commentaires d'un utilisateur
func GetUserComments(userID int) ([]*Comment, error) {
	query := `
		SELECT c.id, c.content, c.user_id, c.post_id, c.parent_comment_id, c.created_at, u.username 
		FROM comment c 
		JOIN user u ON c.user_id = u.id 
		WHERE c.user_id = ? 
		ORDER BY c.created_at DESC`

	rows, err := DbContext.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des commentaires : %v", err)
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		comment := &Comment{}
		var parentID sql.NullInt64
		err := rows.Scan(&comment.ID, &comment.Content, &comment.UserID, &comment.PostID, &parentID, &comment.CreatedAt, &comment.Username)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture des commentaires : %v", err)
		}
		comment.ParentCommentID = parentID
		comments = append(comments, comment)
	}
	return comments, nil
}
