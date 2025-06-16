package main

import (
	"context"
	"database/sql"
	"fmt"
	"forum/config"
	"forum/controllers"
	"forum/models"
	"forum/utils"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DbContext *sql.DB

// InitDB initialise la connexion à la base MySQL et vérifie sa validité
func InitDB() (*sql.DB, error) {
	if DbContext != nil {
		return DbContext, nil
	}

	// Charger la configuration
	if err := config.LoadConfig(); err != nil {
		return nil, fmt.Errorf("erreur lors du chargement de la configuration : %v", err)
	}

	var err error
	DbContext, err = sql.Open("mysql", config.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("échec de l'ouverture de la base de données : %v", err)
	}

	// Vérifie que la connexion est bien établie
	if err = DbContext.Ping(); err != nil {
		return nil, fmt.Errorf("échec de la connexion à la base de données : %v", err)
	}

	return DbContext, nil
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

// withTemplates est un middleware qui ajoute les templates au contexte de la requête
func withTemplates(tmpl *template.Template, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "templates", tmpl)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func main() {
	// Initialisation de la base de données
	db, err := InitDB()
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation de la base de données : %v", err)
	}
	models.DbContext = db
	utils.DbContext = db

	// Création d'un vote positif sur le post 1 par l'utilisateur 1
	voteController := &controllers.VoteController{}
	postID := 1
	err = voteController.CreateVote(1, &postID, nil, 0)
	err = voteController.CreateVote(2, &postID, nil, 1)

	// Initialiser les contrôleurs
	userController := &controllers.UserController{}
	postController := &controllers.PostController{}
	commentController := &controllers.CommentController{}
	searchController := &controllers.SearchController{}
	errorController := &controllers.ErrorController{}

	// Configuration des fonctions du template
	funcMap := template.FuncMap{
		"formatTimeAgo": formatTimeAgo,
	}

	// Chargement des templates avec les fonctions
	templates := template.New("").Funcs(funcMap)
	templates = template.Must(templates.ParseGlob("views/*.html"))

	// Configurer les routes HTTP
	http.HandleFunc("/", errorController.WithErrorHandler(templates, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		postController.IndexHandler(templates)(w, r)
	}))

	http.HandleFunc("/error", errorController.ErrorHandler(templates))
	http.HandleFunc("/login", userController.LoginHandler(templates))
	http.HandleFunc("/register", userController.RegisterHandler(templates))
	http.HandleFunc("/logout", userController.LogoutHandler())
	http.HandleFunc("/create-post", errorController.WithErrorHandler(templates, postController.CreatePostHandler(templates)))
	http.HandleFunc("/post/", errorController.WithErrorHandler(templates, postController.DisplayPostHandler(templates)))
	http.HandleFunc("/comment", errorController.WithErrorHandler(templates, commentController.CreateCommentHandler()))
	http.HandleFunc("/search", errorController.WithErrorHandler(templates, searchController.SearchHandler(templates)))
	http.HandleFunc("/vote", errorController.WithErrorHandler(templates, voteController.VoteHandler()))
	http.HandleFunc("/users/", errorController.WithErrorHandler(templates, userController.ProfileHandler(templates)))

	// Configuration des fichiers statiques
	http.Handle("/styles/", http.StripPrefix("/styles/", http.FileServer(http.Dir("views/static/css/"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("views/static/images/"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("views/static/font/"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("views/static/js/"))))

	fmt.Printf("Site lancé, pour y accéder allez sur: http://%s:%s", config.AppConfig.Host, config.AppConfig.Port)
	if err := http.ListenAndServe(config.AppConfig.Host+":"+config.AppConfig.Port, nil); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur : %v", err)
	}
}
