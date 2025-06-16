package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"forum/models"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

// Clé de session fixe mais sécurisée (32 octets)
var sessionKey = []byte("8x/A?D(G+KbPeShVmYq3s6v9y$B&E)H@")

// Initialisation du store de session avec une clé fixe
var store = sessions.NewCookieStore(sessionKey)

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 jours
		HttpOnly: true,
		Secure:   false, // Mettre à true en production avec HTTPS
		SameSite: http.SameSiteLaxMode,
	}
}

type UserController struct{}

// generateSalt génère un salt aléatoire de 16 octets
func generateSalt() (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("erreur lors de la génération du salt : %v", err)
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

// RegisterUser gère l'inscription d'un nouvel utilisateur
func (uc *UserController) RegisterUser(username, email, password string) error {
	// Vérifier si l'utilisateur existe déjà
	existingUser, err := models.GetUserByUsername(username)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification de l'utilisateur : %v", err)
	}
	if existingUser != nil {
		return errors.New("un utilisateur avec ce nom existe déjà")
	}

	// Générer un salt
	salt, err := generateSalt()
	if err != nil {
		return fmt.Errorf("erreur lors de la génération du salt : %v", err)
	}

	// Combiner le mot de passe avec le salt
	saltedPassword := password + salt

	// Hasher le mot de passe salé
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("erreur lors du hachage du mot de passe : %v", err)
	}

	// Créer le nouvel utilisateur
	user := &models.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		Salt:     salt,
		Role:     "user", // Rôle par défaut
	}

	return user.CreateUser()
}

// LoginUser gère la connexion d'un utilisateur
func (uc *UserController) LoginUser(email, password string) (*models.User, error) {
	user, err := models.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération de l'utilisateur : %v", err)
	}
	if user == nil {
		return nil, errors.New("utilisateur non trouvé")
	}

	// Combiner le mot de passe avec le salt
	saltedPassword := password + user.Salt

	// Vérifier le mot de passe
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(saltedPassword))
	if err != nil {
		return nil, errors.New("mot de passe incorrect")
	}

	return user, nil
}

// UpdateUserProfile met à jour le profil d'un utilisateur
func (uc *UserController) UpdateUserProfile(userID int, email, currentPassword, newPassword string) error {
	user, err := models.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération de l'utilisateur : %v", err)
	}
	if user == nil {
		return errors.New("utilisateur non trouvé")
	}

	// Vérifier le mot de passe actuel si fourni
	if currentPassword != "" {
		saltedCurrentPassword := currentPassword + user.Salt
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(saltedCurrentPassword))
		if err != nil {
			return errors.New("mot de passe actuel incorrect")
		}

		// Hasher le nouveau mot de passe
		if newPassword != "" {
			saltedNewPassword := newPassword + user.Salt
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(saltedNewPassword), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("erreur lors du hachage du nouveau mot de passe : %v", err)
			}
			user.Password = string(hashedPassword)
		}
	}

	// Mettre à jour l'email si fourni
	if email != "" {
		user.Email = email
	}

	return user.UpdateUser()
}

// DeleteUserAccount supprime le compte d'un utilisateur
func (uc *UserController) DeleteUserAccount(userID int, password string) error {
	user, err := models.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération de l'utilisateur : %v", err)
	}
	if user == nil {
		return errors.New("utilisateur non trouvé")
	}

	// Vérifier le mot de passe
	saltedPassword := password + user.Salt
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(saltedPassword))
	if err != nil {
		return errors.New("mot de passe incorrect")
	}

	return user.DeleteUser()
}

// RegisterHandler gère la page d'inscription et le processus d'inscription
func (uc *UserController) RegisterHandler(temp *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			log.Printf("Affichage du formulaire d'inscription pour l'IP: %s", r.RemoteAddr)
			err := temp.ExecuteTemplate(w, "register", nil)
			if err != nil {
				log.Printf("Erreur lors du rendu de la page d'inscription: %v", err)
				http.Error(w, "Erreur lors du rendu de la page", http.StatusInternalServerError)
				return
			}
			return
		}

		if r.Method == "POST" {
			log.Printf("Tentative d'inscription depuis l'IP: %s", r.RemoteAddr)

			// Récupération des données du formulaire
			username := r.FormValue("username")
			email := r.FormValue("email")
			password := r.FormValue("password")
			confirmPassword := r.FormValue("confirm-password")

			// Structure pour stocker les erreurs
			type RegisterError struct {
				Username string
				Email    string
				Password string
				General  string
			}
			errors := RegisterError{}

			// Validation des données
			if username == "" || email == "" || password == "" || confirmPassword == "" {
				errors.General = "Tous les champs sont obligatoires"
			}

			if password != confirmPassword {
				errors.Password = "Les mots de passe ne correspondent pas"
			}

			// Vérifier si l'utilisateur existe déjà
			existingUser, err := models.GetUserByUsername(username)
			if err != nil {
				log.Printf("Erreur lors de la vérification de l'utilisateur: %v", err)
				errors.General = "Une erreur est survenue, veuillez réessayer"
			} else if existingUser != nil {
				errors.Username = "Ce nom d'utilisateur est déjà utilisé"
			}

			// Si des erreurs sont présentes, réafficher le formulaire avec les erreurs
			if errors.General != "" || errors.Username != "" || errors.Email != "" || errors.Password != "" {
				data := struct {
					Username string
					Email    string
					Errors   RegisterError
				}{
					Username: username,
					Email:    email,
					Errors:   errors,
				}
				temp.ExecuteTemplate(w, "register", data)
				return
			}

			// Création de l'utilisateur
			err = uc.RegisterUser(username, email, password)
			if err != nil {
				log.Printf("Échec de l'inscription pour l'utilisateur %s: %v", username, err)
				errors.General = "Une erreur est survenue lors de l'inscription, veuillez réessayer"
				temp.ExecuteTemplate(w, "register", struct {
					Username string
					Email    string
					Errors   RegisterError
				}{
					Username: username,
					Email:    email,
					Errors:   errors,
				})
				return
			}

			log.Printf("Inscription réussie pour l'utilisateur: %s (email: %s)", username, email)
			// Redirection vers la page de connexion en cas de succès
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		log.Printf("Méthode non autorisée (%s) depuis l'IP: %s", r.Method, r.RemoteAddr)
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// LoginHandler gère la page de connexion et le processus de connexion
func (uc *UserController) LoginHandler(temp *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// Effacer tous les cookies de session au chargement de la page
			http.SetCookie(w, &http.Cookie{
				Name:     "session-name",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
			})

			log.Printf("Affichage du formulaire de connexion pour l'IP: %s", r.RemoteAddr)
			data := struct {
				Email string
				Error string
			}{
				Email: "",
				Error: "",
			}
			err := temp.ExecuteTemplate(w, "login", data)
			if err != nil {
				log.Printf("Erreur lors du rendu de la page de connexion: %v", err)
				http.Error(w, "Erreur lors du rendu de la page", http.StatusInternalServerError)
				return
			}
			return
		}

		if r.Method == "POST" {
			log.Printf("Tentative de connexion depuis l'IP: %s", r.RemoteAddr)

			// Récupération des données du formulaire
			email := r.FormValue("email")
			password := r.FormValue("password")

			// Structure pour les données du template
			data := struct {
				Email string
				Error string
			}{
				Email: email,
				Error: "",
			}

			// Validation des données
			if email == "" || password == "" {
				data.Error = "Tous les champs sont obligatoires"
				temp.ExecuteTemplate(w, "login", data)
				return
			}

			// Tentative de connexion
			user, err := uc.LoginUser(email, password)
			if err != nil {
				log.Printf("Erreur de connexion: %v", err)
				data.Error = "Email ou mot de passe incorrect"
				temp.ExecuteTemplate(w, "login", data)
				return
			}

			// Effacer l'ancienne session avant d'en créer une nouvelle
			http.SetCookie(w, &http.Cookie{
				Name:     "session-name",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
			})

			// Création de la session
			session, err := store.Get(r, "session-name")
			if err != nil {
				log.Printf("Erreur lors de la création de la session: %v", err)
				// Réafficher le formulaire de connexion
				data.Error = "Une erreur est survenue, veuillez réessayer"
				temp.ExecuteTemplate(w, "login", data)
				return
			}

			// Stockage des informations de l'utilisateur dans la session
			session.Values["user_id"] = user.ID
			session.Values["username"] = user.Username
			session.Values["role"] = user.Role

			// Sauvegarde de la session
			err = session.Save(r, w)
			if err != nil {
				log.Printf("Erreur lors de la sauvegarde de la session: %v", err)
				http.Error(w, "Erreur serveur", http.StatusInternalServerError)
				return
			}

			// Redirection vers la page d'accueil
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}

// IsConnected vérifie si un utilisateur est connecté et retourne l'utilisateur si c'est le cas
func (uc *UserController) IsConnected(r *http.Request) (*models.User, error) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		return nil, fmt.Errorf("session invalide : %v", err)
	}

	userID, ok := session.Values["user_id"].(int)
	if !ok {
		return nil, nil // Pas d'utilisateur connecté
	}

	user, err := models.GetUserById(models.DbContext, userID)
	if err != nil {
		return nil, fmt.Errorf("utilisateur non trouvé : %v", err)
	}

	return user, nil
}

// ProfileHandler gère l'affichage du profil d'un utilisateur
func (uc *UserController) ProfileHandler(temp *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extraire l'ID de l'URL
		userID, err := strconv.Atoi(r.URL.Path[len("/users/"):])
		if err != nil {
			http.Error(w, "ID utilisateur invalide", http.StatusBadRequest)
			return
		}

		// Récupérer l'utilisateur depuis la base de données
		user, err := models.GetUserByID(userID)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération de l'utilisateur", http.StatusInternalServerError)
			return
		}
		if user == nil {
			http.Error(w, "Utilisateur non trouvé", http.StatusNotFound)
			return
		}

		// Récupérer les statistiques de l'utilisateur
		postController := &PostController{}
		posts, err := postController.GetUserPosts(userID)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
			return
		}

		commentController := &CommentController{}
		comments, err := commentController.GetCommentsByPostID(userID)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des commentaires", http.StatusInternalServerError)
			return
		}

		// Mettre à jour les statistiques
		user.TotalPosts = len(posts)
		user.TotalReplies = len(comments)
		user.Status = "Membre Actif" // À implémenter avec une vraie logique

		// Récupérer l'utilisateur connecté
		currentUser, err := uc.IsConnected(r)
		if err != nil {
			http.Error(w, "Erreur de session", http.StatusInternalServerError)
			return
		}

		data := struct {
			User        *models.User
			CurrentUser *models.User
		}{
			User:        user,
			CurrentUser: currentUser,
		}

		err = temp.ExecuteTemplate(w, "profile", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// LogoutHandler gère la déconnexion d'un utilisateur
func (uc *UserController) LogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Récupérer la session
		session, err := store.Get(r, "session-name")
		if err != nil {
			log.Printf("[ERROR] Erreur lors de la récupération de la session: %v", err)
			http.Error(w, "Erreur lors de la déconnexion", http.StatusInternalServerError)
			return
		}

		// Supprimer les valeurs de la session
		session.Values = make(map[interface{}]interface{})
		session.Options.MaxAge = -1

		// Sauvegarder la session
		err = session.Save(r, w)
		if err != nil {
			log.Printf("[ERROR] Erreur lors de la sauvegarde de la session: %v", err)
			http.Error(w, "Erreur lors de la déconnexion", http.StatusInternalServerError)
			return
		}

		// Rediriger vers la page d'accueil
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
