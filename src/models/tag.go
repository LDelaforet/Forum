package models

import (
	"database/sql"
	"fmt"
)

type Tag struct {
	ID   int
	Name string
}

// GetMaxID récupère le dernier ID utilisé dans la table tag
func GetMaxTagID() (int, error) {
	var maxID int
	query := "SELECT MAX(id) FROM tag"
	err := DbContext.QueryRow(query).Scan(&maxID)
	if err != nil {
		return 0, fmt.Errorf("erreur lors de la récupération du dernier ID : %v", err)
	}
	return maxID, nil
}

// CreateTag crée un nouveau tag
func (t *Tag) CreateTag() error {
	maxID, err := GetMaxTagID()
	if err != nil {
		return err
	}
	t.ID = maxID + 1

	query := "INSERT INTO tag (id, name) VALUES (?, ?)"
	_, err = DbContext.Exec(query, t.ID, t.Name)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du tag : %v", err)
	}
	return nil
}

// GetTagByID récupère un tag par son ID
func GetTagByID(id int) (*Tag, error) {
	tag := &Tag{}
	query := "SELECT id, name FROM tag WHERE id = ?"
	err := DbContext.QueryRow(query, id).Scan(&tag.ID, &tag.Name)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération du tag : %v", err)
	}
	return tag, nil
}

// GetTagByName récupère un tag par son nom
func GetTagByName(name string) (*Tag, error) {
	tag := &Tag{}
	query := "SELECT id, name FROM tag WHERE name = ?"
	err := DbContext.QueryRow(query, name).Scan(&tag.ID, &tag.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			// Le tag n'existe pas, on retourne nil sans erreur
			return nil, nil
		}
		return nil, fmt.Errorf("erreur lors de la récupération du tag : %v", err)
	}
	return tag, nil
}

// GetAllTags récupère tous les tags
func GetAllTags() ([]*Tag, error) {
	query := "SELECT id, name FROM tag ORDER BY name"
	rows, err := DbContext.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des tags : %v", err)
	}
	defer rows.Close()

	var tags []*Tag
	for rows.Next() {
		tag := &Tag{}
		err := rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture des tags : %v", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// AddTagToPost associe un tag à un post
func AddTagToPost(postID, tagID int) error {
	query := "INSERT INTO post_tag (post_id, tag_id) VALUES (?, ?)"
	_, err := DbContext.Exec(query, postID, tagID)
	if err != nil {
		return fmt.Errorf("erreur lors de l'association du tag au post : %v", err)
	}
	return nil
}

// RemoveTagFromPost retire un tag d'un post
func RemoveTagFromPost(postID, tagID int) error {
	query := "DELETE FROM post_tag WHERE post_id = ? AND tag_id = ?"
	_, err := DbContext.Exec(query, postID, tagID)
	if err != nil {
		return fmt.Errorf("erreur lors du retrait du tag du post : %v", err)
	}
	return nil
}

// GetPostTags récupère tous les tags d'un post
func GetPostTags(postID int) ([]*Tag, error) {
	query := `
		SELECT t.id, t.name 
		FROM tag t 
		JOIN post_tag pt ON t.id = pt.tag_id 
		WHERE pt.post_id = ?
		ORDER BY t.name`
	rows, err := DbContext.Query(query, postID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des tags du post : %v", err)
	}
	defer rows.Close()

	var tags []*Tag
	for rows.Next() {
		tag := &Tag{}
		err := rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture des tags : %v", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}
