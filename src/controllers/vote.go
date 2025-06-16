package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"forum/models"
	"io"
	"log"
	"net/http"
)

type VoteController struct{}

// CreateVote crée un nouveau vote
func (vc *VoteController) CreateVote(userID int, postID *int, commentID *int, value int) error {
	// Vérifier si l'utilisateur a déjà voté sur ce post
	existingVote, err := vc.GetVoteByUserAndPost(userID, *postID)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification du vote existant : %v", err)
	}

	// Si un vote existe déjà, on le met à jour
	if existingVote != nil {
		return vc.UpdateVote(existingVote.ID, value)
	}

	// Si aucun vote n'existe, on récupère le prochain ID disponible
	maxId, err := models.GetMaxID("vote")
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération de l'ID maximum : %v", err)
	}

	id := maxId + 1

	if postID != nil {
		_, err = models.DbContext.Exec(
			`INSERT INTO vote (id, user_id, post_id, value, created_at) VALUES (?, ?, ?, ?, NOW())`,
			id, userID, *postID, value)
	} else if commentID != nil {
		_, err = models.DbContext.Exec(
			`INSERT INTO vote (id, user_id, comment_id, value, created_at) VALUES (?, ?, ?, ?, NOW())`,
			id, userID, *commentID, value)
	} else {
		return fmt.Errorf("au moins un des paramètres post_id ou comment_id doit être non-nul")
	}
	if err != nil {
		return fmt.Errorf("échec de la création du vote: %v", err)
	}
	return nil
}

// GetVoteByUserAndPost récupère le vote d'un utilisateur sur un post
func (vc *VoteController) GetVoteByUserAndPost(userID, postID int) (*models.Vote, error) {
	vote := &models.Vote{}
	query := "SELECT id, user_id, post_id, value FROM vote WHERE user_id = ? AND post_id = ?"
	err := models.DbContext.QueryRow(query, userID, postID).Scan(&vote.ID, &vote.UserID, &vote.PostID, &vote.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("erreur lors de la récupération du vote : %v", err)
	}
	return vote, nil
}

// UpdateVote met à jour un vote existant
func (vc *VoteController) UpdateVote(voteID int, value int) error {
	query := "UPDATE vote SET value = ? WHERE id = ?"
	_, err := models.DbContext.Exec(query, value, voteID)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour du vote : %v", err)
	}
	return nil
}

// DeleteVote supprime un vote
func (vc *VoteController) DeleteVote(voteID int) error {
	query := "DELETE FROM vote WHERE id = ?"
	_, err := models.DbContext.Exec(query, voteID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression du vote : %v", err)
	}
	return nil
}

// GetPostVotes récupère tous les votes d'un post
func (vc *VoteController) GetPostVotes(postID int) ([]*models.Vote, error) {
	query := "SELECT id, user_id, post_id, value FROM vote WHERE post_id = ?"
	rows, err := models.DbContext.Query(query, postID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des votes : %v", err)
	}
	defer rows.Close()

	var votes []*models.Vote
	for rows.Next() {
		vote := &models.Vote{}
		err := rows.Scan(&vote.ID, &vote.UserID, &vote.PostID, &vote.Value)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture des votes : %v", err)
		}
		votes = append(votes, vote)
	}
	return votes, nil
}

// GetPostScore récupère le score total des votes pour un post
func (vc *VoteController) GetPostScore(postID int) (int, error) {
	query := "SELECT COALESCE(SUM(value), 0) FROM vote WHERE post_id = ?"
	var score int
	err := models.DbContext.QueryRow(query, postID).Scan(&score)
	if err != nil {
		return 0, fmt.Errorf("erreur lors de la récupération du score du post : %v", err)
	}
	return score, nil
}

// GetUserVotes récupère tous les votes d'un utilisateur
func (vc *VoteController) GetUserVotes(userID int) ([]*models.Vote, error) {
	query := "SELECT id, user_id, post_id, value FROM vote WHERE user_id = ?"
	rows, err := models.DbContext.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des votes : %v", err)
	}
	defer rows.Close()

	var votes []*models.Vote
	for rows.Next() {
		vote := &models.Vote{}
		err := rows.Scan(&vote.ID, &vote.UserID, &vote.PostID, &vote.Value)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture des votes : %v", err)
		}
		votes = append(votes, vote)
	}
	return votes, nil
}

// VoteHandler gère les requêtes de vote
func (vc *VoteController) VoteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Requête de vote reçue: %s %s", r.Method, r.URL.Path)

		if r.Method != http.MethodPost {
			log.Printf("Méthode non autorisée: %s", r.Method)
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		// Vérifier si l'utilisateur est connecté
		userController := &UserController{}
		currentUser, err := userController.IsConnected(r)
		if err != nil {
			response := map[string]interface{}{
				"error":    "Session invalide",
				"redirect": "/login",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
		if currentUser == nil {
			log.Printf("Bah t'es pas co quoi")
			response := map[string]interface{}{
				"error": "Vous devez être connecté pour voter",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Décodage du JSON
		var voteData struct {
			PostID int `json:"post_id"`
			Value  int `json:"value"`
		}

		// Lecture du corps de la requête
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Erreur lors de la lecture du corps de la requête: %v", err)
			http.Error(w, "Erreur lors de la lecture des données", http.StatusBadRequest)
			return
		}
		log.Printf("Corps de la requête reçu: %s", string(body))

		// Décodage du JSON
		if err := json.Unmarshal(body, &voteData); err != nil {
			log.Printf("Erreur lors du décodage des données: %v", err)
			http.Error(w, fmt.Sprintf("Erreur lors du décodage des données: %v", err), http.StatusBadRequest)
			return
		}
		log.Printf("Données de vote reçues: post_id=%d, value=%d", voteData.PostID, voteData.Value)

		// Création du vote avec l'ID de l'utilisateur connecté
		log.Printf("Tentative de création du vote pour user_id=%d, post_id=%d, value=%d", currentUser.ID, voteData.PostID, voteData.Value)
		err = vc.CreateVote(currentUser.ID, &voteData.PostID, nil, voteData.Value)
		if err != nil {
			log.Printf("Erreur lors de la création du vote: %v", err)
			http.Error(w, fmt.Sprintf("Erreur lors de la création du vote : %v", err), http.StatusInternalServerError)
			return
		}
		log.Printf("Vote créé avec succès")

		// Récupération du nouveau score
		log.Printf("Récupération du score pour post_id=%d", voteData.PostID)
		score, err := vc.GetPostScore(voteData.PostID)
		if err != nil {
			log.Printf("Erreur lors de la récupération du score: %v", err)
			http.Error(w, fmt.Sprintf("Erreur lors de la récupération du score : %v", err), http.StatusInternalServerError)
			return
		}
		log.Printf("Score récupéré: %d", score)

		// Envoi de la réponse
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"score":   score,
			"success": true,
		}
		log.Printf("Envoi de la réponse: %v", response)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Erreur lors de l'envoi de la réponse: %v", err)
			http.Error(w, "Erreur lors de l'envoi de la réponse", http.StatusInternalServerError)
			return
		}
	}
}
