package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"forum/models"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PostController struct{}

type PostListVars struct {
	Posts []*models.Post
}

// CreateNewPost crée un nouveau post
func (pc *PostController) CreateNewPost(title, content string, userID int, tags []string) (*models.Post, error) {
	fmt.Printf("[INFO] Début de la création d'un nouveau post - Titre: %s, UserID: %d\n", title, userID)

	if title == "" {
		fmt.Println("[ERROR] Tentative de création de post avec un titre vide")
		return nil, errors.New("le titre ne peut pas être vide")
	}
	if content == "" {
		fmt.Println("[ERROR] Tentative de création de post avec un contenu vide")
		return nil, errors.New("le contenu ne peut pas être vide")
	}

	// Récupérer le dernier ID disponible
	maxId, err := models.GetMaxID("post")
	if err != nil {
		fmt.Printf("[ERROR] Erreur lors de la récupération de l'ID maximum: %v\n", err)
		return nil, fmt.Errorf("erreur lors de la récupération de l'ID maximum : %v", err)
	}
	newId := maxId + 1
	fmt.Printf("[INFO] Nouvel ID généré pour le post: %d\n", newId)

	query := "INSERT INTO post (id, title, content, user_id, created_at) VALUES (?, ?, ?, ?, NOW())"
	_, err = models.DbContext.Exec(query, newId, title, content, userID)
	if err != nil {
		fmt.Printf("[ERROR] Erreur lors de l'insertion du post dans la base de données: %v\n", err)
		return nil, fmt.Errorf("erreur lors de la création du post : %v", err)
	}
	fmt.Printf("[INFO] Post créé avec succès dans la base de données\n")

	post := &models.Post{
		ID:        newId,
		Title:     title,
		Content:   content,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	// Gestion des tags
	fmt.Printf("[INFO] Début du traitement des tags - Nombre de tags: %d\n", len(tags))
	for _, tagName := range tags {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			fmt.Println("[WARN] Tag vide ignoré")
			continue
		}

		fmt.Printf("[INFO] Traitement du tag: %s\n", tagName)

		// Vérifier si le tag existe déjà
		tag, err := models.GetTagByName(tagName)
		if err != nil {
			fmt.Printf("[ERROR] Erreur lors de la vérification du tag '%s': %v\n", tagName, err)
			return nil, fmt.Errorf("erreur lors de la vérification du tag : %v", err)
		}

		if tag == nil {
			fmt.Printf("[INFO] Le tag '%s' n'existe pas, création en cours...\n", tagName)
			// Créer un nouveau tag
			tag = &models.Tag{
				Name: tagName,
			}
			err = tag.CreateTag()
			if err != nil {
				fmt.Printf("[ERROR] Erreur lors de la création du tag '%s': %v\n", tagName, err)
				return nil, fmt.Errorf("erreur lors de la création du tag : %v", err)
			}
			fmt.Printf("[INFO] Nouveau tag créé avec l'ID: %d\n", tag.ID)
		} else {
			fmt.Printf("[INFO] Tag existant trouvé avec l'ID: %d\n", tag.ID)
		}

		// Associer le tag au post
		err = models.AddTagToPost(post.ID, tag.ID)
		if err != nil {
			fmt.Printf("[ERROR] Erreur lors de l'association du tag %d au post %d: %v\n", tag.ID, post.ID, err)
			return nil, fmt.Errorf("erreur lors de l'association du tag au post : %v", err)
		}
		fmt.Printf("[INFO] Tag %d associé avec succès au post %d\n", tag.ID, post.ID)
	}

	fmt.Printf("[INFO] Création du post terminée avec succès - ID: %d\n", post.ID)
	return post, nil
}

// GetPost récupère un post par son ID
func (pc *PostController) GetPost(postID int) (*models.Post, error) {
	query := "SELECT id, title, content, user_id, created_at FROM post WHERE id = ?"
	post := &models.Post{}
	err := models.DbContext.QueryRow(query, postID).Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("post non trouvé")
		}
		return nil, fmt.Errorf("erreur lors de la récupération du post : %v", err)
	}
	return post, nil
}

// UpdatePost met à jour un post existant
func (pc *PostController) UpdatePost(postID, userID int, title, content string) error {
	// Vérifier que l'utilisateur est bien l'auteur du post
	var postUserID int
	err := models.DbContext.QueryRow("SELECT user_id FROM post WHERE id = ?", postID).Scan(&postUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("post non trouvé")
		}
		return fmt.Errorf("erreur lors de la récupération du post : %v", err)
	}

	if postUserID != userID {
		return errors.New("vous n'êtes pas autorisé à modifier ce post")
	}

	// Construire la requête de mise à jour
	query := "UPDATE post SET"
	args := []interface{}{}
	setFields := []string{}

	if title != "" {
		setFields = append(setFields, "title = ?")
		args = append(args, title)
	}
	if content != "" {
		setFields = append(setFields, "content = ?")
		args = append(args, content)
	}

	if len(setFields) == 0 {
		return errors.New("aucune modification à apporter")
	}

	query += " " + (setFields[0])
	for i := 1; i < len(setFields); i++ {
		query += ", " + setFields[i]
	}
	query += " WHERE id = ?"
	args = append(args, postID)

	_, err = models.DbContext.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour du post : %v", err)
	}

	return nil
}

// DeletePost supprime un post
func (pc *PostController) DeletePost(postID, userID int) error {
	// Vérifier que l'utilisateur est bien l'auteur du post
	var postUserID int
	err := models.DbContext.QueryRow("SELECT user_id FROM post WHERE id = ?", postID).Scan(&postUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("post non trouvé")
		}
		return fmt.Errorf("erreur lors de la récupération du post : %v", err)
	}

	if postUserID != userID {
		return errors.New("vous n'êtes pas autorisé à supprimer ce post")
	}

	_, err = models.DbContext.Exec("DELETE FROM post WHERE id = ?", postID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression du post : %v", err)
	}

	return nil
}

// GetUserPosts récupère tous les posts d'un utilisateur
func (pc *PostController) GetUserPosts(userID int) ([]*models.Post, error) {
	query := "SELECT id, title, content, user_id, created_at FROM post WHERE user_id = ? ORDER BY created_at DESC"
	rows, err := models.DbContext.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des posts : %v", err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture des posts : %v", err)
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// GetXPosts récupère un nombre spécifique de posts
func (pc *PostController) GetXPosts(amount int) ([]*models.Post, error) {
	if amount == 0 {
		// Si amount est 0, récupérer tous les posts
		query := "SELECT id, title, content, user_id, created_at FROM post ORDER BY created_at DESC"
		rows, err := models.DbContext.Query(query)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la récupération de tous les posts : %v", err)
		}
		defer rows.Close()

		var posts []*models.Post
		for rows.Next() {
			post := &models.Post{}
			err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt)
			if err != nil {
				return nil, fmt.Errorf("erreur lors de la lecture des posts : %v", err)
			}
			posts = append(posts, post)
		}
		return posts, nil
	}

	// Sinon, récupérer le nombre spécifié de posts
	query := "SELECT id, title, content, user_id, created_at FROM post ORDER BY created_at DESC LIMIT ?"
	rows, err := models.DbContext.Query(query, amount)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des posts : %v", err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture des posts : %v", err)
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// GetAllPosts récupère tous les posts
func (pc *PostController) GetAllPosts() ([]*models.Post, error) {
	query := "SELECT id, title, content, user_id, created_at FROM post ORDER BY created_at DESC"
	rows, err := models.DbContext.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des posts : %v", err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture des posts : %v", err)
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// GetPostsPaginated récupère les posts avec pagination
func (pc *PostController) GetPostsPaginated(page, pageSize int) ([]*models.Post, error) {
	offset := (page - 1) * pageSize
	query := "SELECT id, title, content, user_id, created_at FROM post LIMIT ? OFFSET ?"
	rows, err := models.DbContext.Query(query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des posts paginés : %v", err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture des posts : %v", err)
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// SearchPosts recherche des posts
func (pc *PostController) SearchPosts(query string) ([]*models.Post, error) {
	searchQuery := "SELECT id, title, content, user_id, created_at FROM post WHERE title LIKE ? OR content LIKE ?"
	rows, err := models.DbContext.Query(searchQuery, "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recherche des posts : %v", err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture des posts : %v", err)
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// IsPostOwner vérifie si l'utilisateur est le propriétaire du post
func (pc *PostController) IsPostOwner(postID, userID int) (bool, error) {
	var ownerID int
	err := models.DbContext.QueryRow("SELECT user_id FROM post WHERE id = ?", postID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("post avec ID %d non trouvé", postID)
		}
		return false, fmt.Errorf("erreur lors de la récupération du propriétaire du post : %v", err)
	}
	return ownerID == userID, nil
}

// IndexHandler gère la page d'accueil
func (pc *PostController) IndexHandler(temp *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := pc.GetXPosts(0) // 0 pour récupérer tous les posts
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des posts: "+err.Error(), http.StatusInternalServerError)
			return
		}

		voteController := &VoteController{}

		// Récupération de l'username pour chaque post
		for _, post := range posts {
			user, err := models.GetUserById(models.DbContext, post.UserID)
			if err != nil {
				http.Error(w, "Erreur lors de la récupération de l'utilisateur: "+err.Error(), http.StatusInternalServerError)
				return
			}
			post.Username = user.Username

			voteScore, err := voteController.GetPostScore(post.ID)
			if err != nil {
				http.Error(w, "Erreur lors de la récupération du score: "+err.Error(), http.StatusInternalServerError)
				return
			}
			post.VoteScore = voteScore
		}

		// Récupération de l'utilisateur connecté
		userController := &UserController{}
		currentUser, err := userController.IsConnected(r)
		if err != nil {
			// Si la session est invalide, on redirige vers la page de connexion
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		vars := struct {
			Posts []*models.Post
			User  *models.User
		}{
			Posts: posts,
			User:  currentUser,
		}

		if err := temp.ExecuteTemplate(w, "index", vars); err != nil {
			http.Error(w, "Erreur lors du rendu de la page", http.StatusInternalServerError)
			return
		}
	}
}

// CreatePostHandler gère la création d'un nouveau post
func (pc *PostController) CreatePostHandler(temp *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Vérifier si l'utilisateur est connecté
		userController := &UserController{}
		user, err := userController.IsConnected(r)
		if err != nil {
			// Rediriger vers la page de login si non connecté
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if r.Method == "GET" {
			err := temp.ExecuteTemplate(w, "CreatePost", nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}

		if r.Method == "POST" {
			// Récupérer les données du formulaire
			title := r.FormValue("title")
			content := r.FormValue("content")
			tagsStr := r.FormValue("tags")

			// Traiter les tags
			var tags []string
			if tagsStr != "" {
				tags = strings.Split(tagsStr, ",")
			}

			// Créer le post
			post, err := pc.CreateNewPost(title, content, user.ID, tags)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Rediriger vers la page du post
			http.Redirect(w, r, fmt.Sprintf("/post/%d", post.ID), http.StatusSeeOther)
		}
	}
}

// formatTimeAgo formate le temps écoulé depuis la date donnée
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "À l'instant"
	}
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "Il y a 1 minute"
		}
		return fmt.Sprintf("Il y a %d minutes", minutes)
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "Il y a 1 heure"
		}
		return fmt.Sprintf("Il y a %d heures", hours)
	}
	if duration < 30*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "Il y a 1 jour"
		}
		return fmt.Sprintf("Il y a %d jours", days)
	}
	if duration < 365*24*time.Hour {
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "Il y a 1 mois"
		}
		return fmt.Sprintf("Il y a %d mois", months)
	}
	years := int(duration.Hours() / 24 / 365)
	if years == 1 {
		return "Il y a 1 an"
	}
	return fmt.Sprintf("Il y a %d ans", years)
}

// DisplayPostHandler gère l'affichage d'un post avec ses commentaires
func (pc *PostController) DisplayPostHandler(temp *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[INFO] Début de l'affichage d'un post - URL: %s\n", r.URL.Path)

		// Récupérer l'ID du post depuis l'URL
		postID, err := strconv.Atoi(r.URL.Path[len("/post/"):])
		if err != nil {
			fmt.Printf("[ERROR] ID de post invalide dans l'URL: %s\n", r.URL.Path)
			http.Error(w, "ID de post invalide", http.StatusBadRequest)
			return
		}
		fmt.Printf("[INFO] Affichage du post avec l'ID: %d\n", postID)

		// Récupérer le post
		post, err := pc.GetPost(postID)
		if err != nil {
			fmt.Printf("[ERROR] Erreur lors de la récupération du post %d: %v\n", postID, err)
			http.Error(w, "Post non trouvé", http.StatusNotFound)
			return
		}
		fmt.Printf("[INFO] Post %d récupéré avec succès\n", postID)

		// Récupérer l'utilisateur connecté (sans redirection si non connecté)
		userController := &UserController{}
		currentUser, _ := userController.IsConnected(r)
		if currentUser != nil {
			fmt.Printf("[INFO] Utilisateur connecté: %s (ID: %d)\n", currentUser.Username, currentUser.ID)
		} else {
			fmt.Printf("[INFO] Aucun utilisateur connecté\n")
		}

		// Récupérer les commentaires du post
		commentController := &CommentController{}
		comments, err := commentController.GetCommentsByPostID(postID)
		if err != nil {
			fmt.Printf("[ERROR] Erreur lors de la récupération des commentaires du post %d: %v\n", postID, err)
			http.Error(w, "Erreur lors de la récupération des commentaires", http.StatusInternalServerError)
			return
		}
		fmt.Printf("[INFO] %d commentaires récupérés pour le post %d\n", len(comments), postID)

		// Récupérer le score du post
		voteController := &VoteController{}
		voteScore, err := voteController.GetPostScore(postID)
		if err != nil {
			fmt.Printf("[ERROR] Erreur lors de la récupération du score du post %d: %v\n", postID, err)
			http.Error(w, "Erreur lors de la récupération du score", http.StatusInternalServerError)
			return
		}
		fmt.Printf("[INFO] Score du post %d: %d\n", postID, voteScore)
		post.VoteScore = voteScore

		// Récupérer le nom d'utilisateur de l'auteur du post
		author, err := models.GetUserById(models.DbContext, post.UserID)
		if err != nil {
			fmt.Printf("[ERROR] Erreur lors de la récupération de l'auteur du post %d: %v\n", postID, err)
			http.Error(w, "Erreur lors de la récupération de l'auteur", http.StatusInternalServerError)
			return
		}
		fmt.Printf("[INFO] Auteur du post %d: %s (ID: %d)\n", postID, author.Username, author.ID)
		post.Username = author.Username

		// Récupérer les tags du post
		tags, err := models.GetPostTags(postID)
		if err != nil {
			fmt.Printf("[ERROR] Erreur lors de la récupération des tags du post %d: %v\n", postID, err)
			http.Error(w, "Erreur lors de la récupération des tags", http.StatusInternalServerError)
			return
		}
		fmt.Printf("[INFO] Tags récupérés pour le post %d: %v\n", postID, tags)

		// Ajouter la fonction formatTimeAgo au template
		funcMap := template.FuncMap{
			"formatTimeAgo": formatTimeAgo,
		}
		temp = temp.Funcs(funcMap)

		vars := struct {
			Post     *models.Post
			Comments []*models.Comment
			User     *models.User
			Tags     []*models.Tag
		}{
			Post:     post,
			Comments: comments,
			User:     currentUser,
			Tags:     tags,
		}

		if err := temp.ExecuteTemplate(w, "post", vars); err != nil {
			fmt.Printf("[ERROR] Erreur lors du rendu de la page du post %d: %v\n", postID, err)
			http.Error(w, "Erreur lors du rendu de la page", http.StatusInternalServerError)
			return
		}
		fmt.Printf("[INFO] Affichage du post %d terminé avec succès\n", postID)
	}
}
