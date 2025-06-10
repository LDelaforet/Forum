package utils

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql" // Import du driver MySQL (anonyme)
)

var DbContext *sql.DB

func GetEnvWithDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

// InitDB initialise la connexion à la base MySQL et vérifie sa validité
func InitDB() (*sql.DB, error) {
	if DbContext != nil {
		log.Println("La connexion à la base de données est déjà initialisée.")
		return DbContext, nil
	}
	user := GetEnvWithDefault("DB_USER", "root")
	host := GetEnvWithDefault("DB_HOST", "localhost")
	port := GetEnvWithDefault("DB_PORT", "3306")
	name := GetEnvWithDefault("DB_NAME", "bdd_forum")

	if user == "" || name == "" {
		return nil, fmt.Errorf("les variables d'environnement DB_USER et DB_NAME doivent être définies")
	}

	// Format DSN pour MySQL : user:password@tcp(host:port)/dbname?parseTime=true
	dsn := fmt.Sprintf("%s:@tcp(%s:%s)/%s?parseTime=true", user, host, port, name)

	var err error
	DbContext, err = sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("échec de l'ouverture de la base de données : %v", err)
	}

	// Vérifie que la connexion est bien établie
	if err = DbContext.Ping(); err != nil {
		return nil, fmt.Errorf("échec de la connexion à la base de données : %v", err)
	}

	log.Println("Connexion à la base de données réussie.")
	return DbContext, nil
}

// GetMaxID récupère le plus grand id de la table passée en paramètre.
func GetMaxID(tableName string) (int, error) {
	var maxID sql.NullInt64
	query := fmt.Sprintf("SELECT MAX(id) FROM %s", tableName)
	err := DbContext.QueryRow(query).Scan(&maxID)
	if err != nil {
		return 0, fmt.Errorf("erreur lors de la récupération du max id : %v", err)
	}
	if !maxID.Valid {
		return 0, nil // table vide => max id = 0
	}
	return int(maxID.Int64), nil
}

// Récupère l'id d'une ligne dans une table en fonction d'une colonne et de sa valeur.
// J'ai fait cette fonction psq j'ai remarqué que je faisait bcp de fois des fonctions très similaires
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

func EditSomethingById(tableName, columnName, value string, id int) error {
	query := fmt.Sprintf("UPDATE %s SET %s = ? WHERE id = ?", tableName, columnName)
	_, err := DbContext.Exec(query, value, id)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour de %s : %v", tableName, err)
	}
	return nil
}

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

func GetCountBySomething(tableName, columnName, value string) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ?", tableName, columnName)
	row := DbContext.QueryRow(query, value)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("erreur lors de la récupération du compte : %v", err)
	}
	return count, nil
}
