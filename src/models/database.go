package models

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DbContext *sql.DB

func GetEnvWithDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

// GetMaxID récupère le plus grand id de la table passée en paramètre.
func GetMaxID(tableName string) (int, error) {
	var maxID sql.NullInt64
	query := fmt.Sprintf("SELECT COALESCE(MAX(id), 0) FROM %s", tableName)
	err := DbContext.QueryRow(query).Scan(&maxID)
	if err != nil {
		return 0, fmt.Errorf("erreur lors de la récupération du max id : %v", err)
	}
	if !maxID.Valid {
		return 0, nil // table vide => max id = 0
	}
	return int(maxID.Int64), nil
}

// GetIdBySomething récupère l'id d'une ligne dans une table en fonction d'une colonne et de sa valeur.
func GetIdBySomething(tableName, columnName, value string) (int, error) {
	var id int
	query := fmt.Sprintf("SELECT id FROM %s WHERE %s = ?", tableName, columnName)
	err := DbContext.QueryRow(query, value).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // Pas de ligne trouvée
		}
		return 0, fmt.Errorf("erreur lors de la récupération de l'id : %v", err)
	}
	return id, nil
}

// GetAllIdsBySomething récupère tous les ids correspondant à un critère
func GetAllIdsBySomething(tableName, columnName, value string) ([]int, error) {
	query := fmt.Sprintf("SELECT id FROM %s WHERE %s = ?", tableName, columnName)
	rows, err := DbContext.Query(query, value)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des ids : %v", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture des lignes : %v", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors de l'itération sur les lignes : %v", err)
	}

	return ids, nil
}

// EditSomethingById modifie une valeur dans une table
func EditSomethingById(tableName, columnName, value string, id int) error {
	query := fmt.Sprintf("UPDATE %s SET %s = ? WHERE id = ?", tableName, columnName)
	_, err := DbContext.Exec(query, value, id)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour de %s : %v", tableName, err)
	}
	return nil
}

// DeleteSomethingById supprime une ligne dans une table
func DeleteSomethingById(tableName string, id int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)
	result, err := DbContext.Exec(query, id)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression de %s : %v", tableName, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération du nombre de lignes affectées : %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("aucune ligne trouvée avec l'id %d dans la table %s", id, tableName)
	}

	return nil
}

// GetAllIds récupère tous les ids d'une table
func GetAllIds(tableName string) ([]int, error) {
	query := fmt.Sprintf("SELECT id FROM %s", tableName)
	rows, err := DbContext.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des ids : %v", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture des lignes : %v", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors de l'itération sur les lignes : %v", err)
	}

	return ids, nil
}

// GetCountBySomething compte le nombre d'éléments correspondant à un critère
func GetCountBySomething(tableName, columnName, value string) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ?", tableName, columnName)
	row := DbContext.QueryRow(query, value)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("erreur lors de la récupération du compte : %v", err)
	}
	return count, nil
}
