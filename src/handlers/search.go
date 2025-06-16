package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strings"
	"time"
)

type SearchResult struct {
	ID         int
	Title      string
	Author     string
	Date       string
	ReplyCount int
	Tags       []string
}

type SearchData struct {
	Query   string
	Type    string
	Results []SearchResult
	User    *User
}

type User struct {
	ID       int
	Username string
}

func SearchHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		searchType := r.URL.Query().Get("type")
		if searchType == "" {
			searchType = "title"
		}

		var results []SearchResult
		var err error

		if query != "" {
			if searchType == "tag" {
				results, err = searchByTag(db, query)
			} else {
				results, err = searchByTitle(db, query)
			}

			if err != nil {
				http.Error(w, "Erreur lors de la recherche: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Récupérer les informations de l'utilisateur depuis la session
		var user *User
		session, err := r.Cookie("session")
		if err == nil {
			// Vérifier la session dans la base de données
			var userID int
			var username string
			err = db.QueryRow("SELECT user_id, username FROM sessions s JOIN user u ON s.user_id = u.id WHERE s.token = ?", session.Value).Scan(&userID, &username)
			if err == nil {
				user = &User{
					ID:       userID,
					Username: username,
				}
			}
		}

		data := SearchData{
			Query:   query,
			Type:    searchType,
			Results: results,
			User:    user,
		}

		tmpl, err := template.ParseFiles("views/search.html")
		if err != nil {
			http.Error(w, "Erreur lors du chargement du template: "+err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Erreur lors de l'affichage de la page: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func searchByTitle(db *sql.DB, query string) ([]SearchResult, error) {
	query = "%" + strings.ToLower(query) + "%"

	rows, err := db.Query(`
		SELECT p.id, p.title, u.username, p.created_at, 
		       (SELECT COUNT(*) FROM comment WHERE post_id = p.id) as reply_count,
		       GROUP_CONCAT(t.name) as tags
		FROM post p
		LEFT JOIN user u ON p.user_id = u.id
		LEFT JOIN post_tag pt ON p.id = pt.post_id
		LEFT JOIN tag t ON pt.tag_id = t.id
		WHERE LOWER(p.title) LIKE ?
		GROUP BY p.id
		ORDER BY p.created_at DESC
	`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		var tags sql.NullString
		var createdAt time.Time

		err := rows.Scan(&result.ID, &result.Title, &result.Author, &createdAt, &result.ReplyCount, &tags)
		if err != nil {
			return nil, err
		}

		result.Date = createdAt.Format("02/01/2006 à 15:04")
		if tags.Valid {
			result.Tags = strings.Split(tags.String, ",")
		}

		results = append(results, result)
	}

	return results, nil
}

func searchByTag(db *sql.DB, query string) ([]SearchResult, error) {
	query = "%" + strings.ToLower(query) + "%"

	rows, err := db.Query(`
		SELECT p.id, p.title, u.username, p.created_at, 
		       (SELECT COUNT(*) FROM comment WHERE post_id = p.id) as reply_count,
		       GROUP_CONCAT(t.name) as tags
		FROM post p
		LEFT JOIN user u ON p.user_id = u.id
		LEFT JOIN post_tag pt ON p.id = pt.post_id
		LEFT JOIN tag t ON pt.tag_id = t.id
		WHERE LOWER(t.name) LIKE ?
		GROUP BY p.id
		ORDER BY p.created_at DESC
	`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		var tags sql.NullString
		var createdAt time.Time

		err := rows.Scan(&result.ID, &result.Title, &result.Author, &createdAt, &result.ReplyCount, &tags)
		if err != nil {
			return nil, err
		}

		result.Date = createdAt.Format("02/01/2006 à 15:04")
		if tags.Valid {
			result.Tags = strings.Split(tags.String, ",")
		}

		results = append(results, result)
	}

	return results, nil
}
