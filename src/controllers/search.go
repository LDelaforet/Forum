package controllers

import (
	"fmt"
	"forum/models"
	"html/template"
	"net/http"
)

type SearchController struct{}

// SearchPosts recherche des posts par titre, contenu ou tags
func (sc *SearchController) SearchPosts(query string, searchType string) ([]*models.Post, error) {
	// Si aucun critère de recherche n'est fourni, retourner tous les posts
	if query == "" {
		postController := &PostController{}
		return postController.GetAllPosts()
	}

	// Construire la requête SQL de base
	baseQuery := `
		SELECT DISTINCT p.id, p.title, p.content, p.user_id, p.created_at 
		FROM post p
		LEFT JOIN post_tag pt ON p.id = pt.post_id
		LEFT JOIN tag t ON pt.tag_id = t.id
		WHERE 1=1
	`
	args := []interface{}{}

	// Ajouter les conditions de recherche selon le type
	if searchType == "tag" {
		// Vérifier si le tag existe
		var tagCount int
		err := models.DbContext.QueryRow("SELECT COUNT(*) FROM tag WHERE name COLLATE utf8mb4_general_ci LIKE ?", "%"+query+"%").Scan(&tagCount)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la vérification de l'existence du tag : %v", err)
		}

		// Si le tag existe mais n'a pas de relations, vérifier s'il y a des posts avec ce mot dans le titre
		if tagCount > 0 {
			var relationCount int
			err = models.DbContext.QueryRow(`
				SELECT COUNT(*) 
				FROM post_tag pt 
				JOIN tag t ON pt.tag_id = t.id 
				WHERE t.name COLLATE utf8mb4_general_ci LIKE ?`, "%"+query+"%").Scan(&relationCount)
			if err != nil {
				return nil, fmt.Errorf("erreur lors de la vérification des relations post_tag : %v", err)
			}

			// Si pas de relations mais le tag existe, chercher des posts avec ce mot dans le titre
			if relationCount == 0 {
				var postCount int
				err = models.DbContext.QueryRow(`
					SELECT COUNT(*) 
					FROM post 
					WHERE title COLLATE utf8mb4_general_ci LIKE ?`, "%"+query+"%").Scan(&postCount)
				if err != nil {
					return nil, fmt.Errorf("erreur lors de la recherche de posts correspondants : %v", err)
				}
			}
		}

		baseQuery = `
			SELECT DISTINCT p.id, p.title, p.content, p.user_id, p.created_at 
			FROM tag t
			INNER JOIN post_tag pt ON t.id = pt.tag_id
			INNER JOIN post p ON pt.post_id = p.id
			WHERE t.name COLLATE utf8mb4_general_ci LIKE ?
		`
		args = append(args, "%"+query+"%")
	} else {
		baseQuery += " AND (p.title COLLATE utf8mb4_general_ci LIKE ? OR p.content COLLATE utf8mb4_general_ci LIKE ?)"
		args = append(args, "%"+query+"%", "%"+query+"%")
	}

	// Ajouter le tri à la fin de la requête
	baseQuery += " ORDER BY p.created_at DESC"

	// Exécuter la requête
	rows, err := models.DbContext.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recherche des posts : %v", err)
	}
	defer rows.Close()

	var posts []*models.Post
	postCount := 0
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture des posts : %v", err)
		}

		// Récupérer le nom d'utilisateur
		user, err := models.GetUserById(models.DbContext, post.UserID)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la récupération de l'utilisateur : %v", err)
		}
		post.Username = user.Username

		// Récupérer le score du post
		voteController := &VoteController{}
		score, err := voteController.GetPostScore(post.ID)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la récupération du score : %v", err)
		}
		post.VoteScore = score

		posts = append(posts, post)
		postCount++
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors de la lecture des posts : %v", err)
	}

	return posts, nil
}

// SearchHandler gère la page de recherche
func (sc *SearchController) SearchHandler(temp *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Récupérer l'utilisateur connecté
		userController := &UserController{}
		currentUser, err := userController.IsConnected(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if r.Method == "GET" {
			// Récupérer les paramètres de recherche
			query := r.URL.Query().Get("query")
			searchType := r.URL.Query().Get("type")
			if searchType == "" {
				searchType = "text"
			}

			// Effectuer la recherche
			posts, err := sc.SearchPosts(query, searchType)
			if err != nil {
				http.Error(w, "Erreur lors de la recherche", http.StatusInternalServerError)
				return
			}

			// Préparer les données pour le template
			vars := struct {
				Posts []*models.Post
				User  *models.User
				Query string
				Type  string
			}{
				Posts: posts,
				User:  currentUser,
				Query: query,
				Type:  searchType,
			}

			// Rendre le template
			if err := temp.ExecuteTemplate(w, "search", vars); err != nil {
				http.Error(w, "Erreur lors du rendu de la page", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	}
}
