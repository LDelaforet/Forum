package controllers

import (
	"forum/models"
	"log"
	"net/http"
	"strconv"
)

type CommentController struct{}

// CreateComment crée un nouveau commentaire
func (cc *CommentController) CreateComment(content string, userID, postID int, parentCommentID *int) error {
	log.Printf("[INFO] Création d'un nouveau commentaire - Content: %s, UserID: %d, PostID: %d, ParentCommentID: %v", content, userID, postID, parentCommentID)
	return models.CreateComment(models.DbContext, content, postID, parentCommentID, userID)
}

// GetCommentsByPostID récupère tous les commentaires d'un post
func (cc *CommentController) GetCommentsByPostID(postID int) ([]*models.Comment, error) {
	log.Printf("[INFO] Récupération des commentaires pour le post %d", postID)
	return models.GetCommentsByPostID(postID)
}

// CreateCommentHandler gère la création d'un nouveau commentaire
func (cc *CommentController) CreateCommentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[INFO] Requête de création de commentaire reçue - Méthode: %s", r.Method)

		if r.Method != "POST" {
			log.Printf("[ERROR] Méthode non autorisée: %s", r.Method)
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		// Récupérer l'ID de l'utilisateur depuis la session
		userController := &UserController{}
		user, err := userController.IsConnected(r)
		if err != nil {
			log.Printf("[ERROR] Erreur de session: %v", err)
			http.Error(w, "Non autorisé", http.StatusUnauthorized)
			return
		}
		log.Printf("[INFO] Utilisateur connecté: %s (ID: %d)", user.Username, user.ID)

		// Récupérer les données du formulaire
		content := r.FormValue("content")
		postIDStr := r.FormValue("post_id")
		parentCommentIDStr := r.FormValue("parent_comment_id")

		log.Printf("[INFO] Données reçues - Content: %s, PostID: %s, ParentCommentID: %s", content, postIDStr, parentCommentIDStr)

		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			log.Printf("[ERROR] ID de post invalide: %s", postIDStr)
			http.Error(w, "ID de post invalide", http.StatusBadRequest)
			return
		}

		var parentCommentID *int
		if parentCommentIDStr != "" {
			parentID, err := strconv.Atoi(parentCommentIDStr)
			if err != nil {
				log.Printf("[ERROR] ID de commentaire parent invalide: %s", parentCommentIDStr)
				http.Error(w, "ID de commentaire parent invalide", http.StatusBadRequest)
				return
			}
			parentCommentID = &parentID
		}

		// Créer le commentaire
		err = cc.CreateComment(content, user.ID, postID, parentCommentID)
		if err != nil {
			log.Printf("[ERROR] Erreur lors de la création du commentaire: %v", err)
			http.Error(w, "Erreur lors de la création du commentaire", http.StatusInternalServerError)
			return
		}

		log.Printf("[INFO] Commentaire créé avec succès pour le post %d", postID)

		// Rediriger vers la page du post
		http.Redirect(w, r, "/post/"+strconv.Itoa(postID), http.StatusSeeOther)
	}
}
