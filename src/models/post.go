package models

import (
	"database/sql"
	"fmt"
	"forum/utils"
	"time"
)

type Post struct {
	ID        int
	Title     string
	Content   string
	UserID    int
	CreatedAt time.Time
}

// GetPostById récupère un post par son ID
func GetPostById(db *sql.DB, id int) (*Post, error) {
	row := db.QueryRow("SELECT id, title, content, user_id, created_at FROM post WHERE id = ?", id)

	var post Post
	err := row.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("échec de la récupération du post : %v", err)
	}
	return &post, nil
}

func CreatePost(db *sql.DB, title string, content string, user_id int) error {
	maxId, err := utils.GetMaxID("post")
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération de l'ID maximum : %v", err)
	}

	id := maxId + 1

	_, err = db.Exec(
		`INSERT INTO post (id, title, content, user_id, created_at) VALUES (?, ?, ?, ?, NOW())`,
		id, title, content, user_id)
	if err != nil {
		return fmt.Errorf("échec de la création du post: %v", err)
	}
	return nil
}

func GetAllPosts(db *sql.DB) ([]Post, error) {
	ids, err := utils.GetAllIdsBySomething("post", "user_id", "")
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des IDs des posts : %v", err)
	}
	posts := make([]Post, 0, len(ids))
	for _, id := range ids {
		post, err := GetPostById(db, id)
		if err != nil {
			return nil, fmt.Errorf("échec de la récupération du post avec ID %d : %v", id, err)
		}
		if post != nil {
			posts = append(posts, *post)
		}
	}
	return posts, nil
}

func GetAllPostsByUserId(db *sql.DB, userId int) ([]Post, error) {
	ids, err := utils.GetAllIdsBySomething("post", "user_id", fmt.Sprintf("%d", userId))
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des IDs des posts pour l'utilisateur %d : %v", userId, err)
	}
	posts := make([]Post, 0, len(ids))
	for _, id := range ids {
		post, err := GetPostById(db, id)
		if err != nil {
			return nil, fmt.Errorf("échec de la récupération du post avec ID %d : %v", id, err)
		}
		if post != nil {
			posts = append(posts, *post)
		}
	}
	return posts, nil
}

func EditPostTitle(db *sql.DB, id int, newTitle string) error {
	err := utils.EditSomethingById("post", "title", newTitle, id)
	if err != nil {
		return fmt.Errorf("échec de la mise à jour du titre du post avec ID %d : %v", id, err)
	}
	return nil
}

func EditPostContent(db *sql.DB, id int, newContent string) error {
	err := utils.EditSomethingById("post", "content", newContent, id)
	if err != nil {
		return fmt.Errorf("échec de la mise à jour du contenu du post avec ID %d : %v", id, err)
	}
	return nil
}

func GetPostVoteScore(db *sql.DB, postId int) (int, error) {
	query := `
		SELECT SUM(value) FROM vote WHERE post_id = ?`
	row := db.QueryRow(query, postId)

	var score sql.NullInt64
	err := row.Scan(&score)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // Pas de votes trouvés
		}
		return 0, fmt.Errorf("échec de la récupération du score de vote pour le post %d : %v", postId, err)
	}

	if score.Valid {
		return int(score.Int64), nil
	}
	return 0, nil // Aucun vote trouvé
}

func DeletePost(db *sql.DB, id int) error {
	err := utils.DeleteSomethingById("post", id)
	if err != nil {
		return fmt.Errorf("échec de la suppression du post avec ID %d : %v", id, err)
	}
	return nil
}

func SortByDate(posts []Post) []Post {
	// Simple bubble sort pour trier les posts par date de création
	for i := 0; i < len(posts)-1; i++ {
		for j := 0; j < len(posts)-i-1; j++ {
			if posts[j].CreatedAt.Before(posts[j+1].CreatedAt) {
				posts[j], posts[j+1] = posts[j+1], posts[j]
			}
		}
	}
	return posts
}

func SortByScore(posts []Post, db *sql.DB) ([]Post, error) {
	// Récupérer les scores de vote pour chaque post
	scores := make(map[int]int)
	for _, post := range posts {
		score, err := GetPostVoteScore(db, post.ID)
		if err != nil {
			return nil, fmt.Errorf("échec de la récupération du score de vote pour le post %d : %v", post.ID, err)
		}
		scores[post.ID] = score
	}

	// Simple bubble sort pour trier les posts par score de vote
	for i := 0; i < len(posts)-1; i++ {
		for j := 0; j < len(posts)-i-1; j++ {
			if scores[posts[j].ID] < scores[posts[j+1].ID] {
				posts[j], posts[j+1] = posts[j+1], posts[j]
			}
		}
	}
	return posts, nil
}

func SortByTag(posts []Post, tag string) []Post {
	// Filtrer les posts par tag
	filteredPosts := make([]Post, 0)
	for _, post := range posts {
		if post.Title == tag || post.Content == tag { // Supposons que le tag soit dans le titre ou le contenu
			filteredPosts = append(filteredPosts, post)
		}
	}
	return filteredPosts
}

func GetPostByTag(db *sql.DB, tag string) ([]Post, error) {
	// Récupérer tous les posts
	posts, err := GetAllPosts(db)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des posts : %v", err)
	}

	// Filtrer les posts par tag
	filteredPosts := SortByTag(posts, tag)
	return filteredPosts, nil
}

// AddTagsToPost associe une liste de tags à un post (dans post_tag)
func AddTagsToPost(db *sql.DB, postID int, tags []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, tagName := range tags {
		// Insère le tag s’il n’existe pas
		res, err := tx.Exec("INSERT IGNORE INTO tag (name) VALUES (?)", tagName)
		if err != nil {
			return err
		}

		// Récupère l'ID du tag
		var tagID int64
		if id, err := res.LastInsertId(); err == nil && id != 0 {
			tagID = id
		} else {
			// Si le tag existait déjà, récupère son ID
			err = tx.QueryRow("SELECT id FROM tag WHERE name = ?", tagName).Scan(&tagID)
			if err != nil {
				return err
			}
		}

		// Insère la relation post-tag
		_, err = tx.Exec("INSERT IGNORE INTO post_tag (post_id, tag_id) VALUES (?, ?)", postID, tagID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetTagsByPostID récupère tous les noms de tags associés à un post
func GetTagsByPostID(db *sql.DB, postID int) ([]string, error) {
	rows, err := db.Query(`
		SELECT t.name 
		FROM tag t 
		JOIN post_tag pt ON t.id = pt.tag_id 
		WHERE pt.post_id = ?`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tagName string
		if err := rows.Scan(&tagName); err != nil {
			return nil, err
		}
		tags = append(tags, tagName)
	}
	return tags, nil
}

func GetPostsPaginated(db *sql.DB, page int, pageSize int) ([]Post, error) {
	offset := (page - 1) * pageSize
	rows, err := db.Query("SELECT id, title, content, user_id, created_at FROM post LIMIT ? OFFSET ?", pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des posts paginés : %v", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt); err != nil {
			return nil, fmt.Errorf("échec de la lecture du post : %v", err)
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func SearchPosts(db *sql.DB, query string) ([]Post, error) {
	rows, err := db.Query("SELECT id, title, content, user_id, created_at FROM post WHERE title LIKE ? OR content LIKE ?", "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("échec de la recherche des posts : %v", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.CreatedAt); err != nil {
			return nil, fmt.Errorf("échec de la lecture du post : %v", err)
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func IsPostOwner(db *sql.DB, postID int, userID int) (bool, error) {
	var ownerID int
	err := db.QueryRow("SELECT user_id FROM post WHERE id = ?", postID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("post avec ID %d non trouvé", postID)
		}
		return false, fmt.Errorf("échec de la récupération du propriétaire du post : %v", err)
	}
	return ownerID == userID, nil
}
