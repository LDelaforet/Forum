package models

import (
	"database/sql"
)

type Tag struct {
	ID   int
	Name string
}

// CreateTag insère un nouveau tag dans la BDD, ignore si existe déjà
func CreateTag(db *sql.DB, name string) (int, error) {
	res, err := db.Exec("INSERT IGNORE INTO tag (name) VALUES (?)", name)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// CreateTagIfNotExists insère un nouveau tag dans la BDD, ignore si existe déjà
func CreateTagIfNotExists(db *sql.DB, tag string) error {
	_, err := db.Exec("INSERT IGNORE INTO tag (name) VALUES (?)", tag)
	return err
}

// GetTagByName récupère un tag par son nom
func GetTagByName(db *sql.DB, name string) (*Tag, error) {
	tag := &Tag{}
	err := db.QueryRow("SELECT id, name FROM tag WHERE name = ?", name).Scan(&tag.ID, &tag.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // tag non trouvé
		}
		return nil, err
	}
	return tag, nil
}

// GetAllTags récupère tous les tags
func GetAllTags(db *sql.DB) ([]Tag, error) {
	rows, err := db.Query("SELECT id, name FROM tag")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}
