package forum

import (
	"fmt"
	"forum/models"
	"forum/utils"
	"testing"

	_ "github.com/go-sql-driver/mysql" // driver mysql
)

func TestBasicFunctions(t *testing.T) {
	// Idéalement, crée une transaction pour rollback après test
	db, err := utils.InitDB()
	if err != nil {
		t.Fatalf("erreur démarrage transaction : %v\n", err)
	}

	// 1. CreateUser
	err = models.CreateUser(db, "testuser", "testuser@example.com", "Motdepasse!123")

	if err != nil {
		fmt.Printf("CreateUser failed: %v\n", err)
	}

	// 2. GetUserById
	user, err := models.GetUserById(db, 1)
	if err != nil {
		fmt.Printf("GetUserById failed: %v\n", err)
	}
	if user == nil {
		t.Error("GetUserById returned nil user")
	}

	// 3. Create a post
	err = models.CreatePost(db, "Titre Test", "Contenu test", 1)
	if err != nil {
		fmt.Printf("CreatePost failed: %v\n", err)
	}

	// 4. GetPostVoteScore
	_, err = models.GetPostVoteScore(db, 1)
	if err != nil {
		fmt.Printf("GetPostVoteScore failed: %v\n", err)
	}

	// 4.1 S'assurer que les tags existent
	for _, tag := range []string{"tag1", "tag2"} {
		err = models.CreateTagIfNotExists(db, tag)
		if err != nil {
			fmt.Printf("CreateTagIfNotExists failed for %s: %v", tag, err)
		}
	}

	// 5. AddTagsToPost
	err = models.AddTagsToPost(db, 1, []string{"tag1", "tag2"})
	if err != nil {
		fmt.Printf("AddTagsToPost failed: %v\n", err)
	}

	// 6. GetTagsByPostID
	_, err = models.GetTagsByPostID(db, 1)
	if err != nil {
		fmt.Printf("GetTagsByPostID failed: %v\n", err)
	}

	// 7. GetPostByTag
	_, err = models.GetPostByTag(db, "tag1")
	if err != nil {
		fmt.Printf("GetPostByTag failed: %v\n", err)
	}

	// 8. SortByDate (c’est juste du trie en mémoire)
	posts := []models.Post{}
	_ = models.SortByDate(posts)

	// 9. SortByScore (peut retourner erreur)
	_, err = models.SortByScore(posts, db)
	if err != nil {
		fmt.Printf("SortByScore failed: %v\n", err)
	}

	// 10. SortByTag (juste filtre mémoire)
	_ = models.SortByTag(posts, "tag1")

	// 11. GetPostsPaginated
	_, err = models.GetPostsPaginated(db, 1, 10)
	if err != nil {
		fmt.Printf("GetPostsPaginated failed: %v\n", err)
	}

	// 12. SearchPosts
	_, err = models.SearchPosts(db, "test")
	if err != nil {
		fmt.Printf("SearchPosts failed: %v\n", err)
	}

	// 13. IsPostOwner
	_, err = models.IsPostOwner(db, 1, 1)
	if err != nil {
		fmt.Printf("IsPostOwner failed: %v\n", err)
	}

	// 14 à 18. Édition (EditPassword, EditEmail, EditUsername)
	if err := models.EditPassword(db, 1, "NouveauPass!123"); err != nil {
		fmt.Printf("EditPassword failed: %v\n", err)
	}
	if err := models.EditEmail(db, 1, "newemail@example.com"); err != nil {
		fmt.Printf("EditEmail failed: %v\n", err)
	}
	if err := models.EditUsername(db, 1, "nouveluser"); err != nil {
		fmt.Printf("EditUsername failed: %v\n", err)
	}

	// Optionnel: Commit la transaction si tu veux garder les données
	// db.Commit()
}

func main() {
	TestBasicFunctions(nil)
}
