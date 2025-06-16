package models

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"forum/utils"
	"strings"
	"time"
)

type User struct {
	ID               int       `json:"id"`
	Username         string    `json:"username"`
	Email            string    `json:"email"`
	Password         string    `json:"-"` // Le mot de passe n'est jamais exposé en JSON
	Salt             string    `json:"-"`
	Role             string    `json:"role"`
	Status           string    `json:"status"`
	RegistrationDate time.Time `json:"registration_date"`
	TotalPosts       int       `json:"total_posts"`
	TotalReplies     int       `json:"total_replies"`
	CreatedAt        time.Time `json:"created_at"`
}

// CreateUser crée un nouvel utilisateur dans la base de données
func (u *User) CreateUser() error {
	// Récupérer le dernier ID
	var maxID int
	err := DbContext.QueryRow("SELECT COALESCE(MAX(id), 0) FROM user").Scan(&maxID)
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération du dernier ID : %v", err)
	}

	query := "INSERT INTO user (id, username, email, passwd_hash, salt, created_at) VALUES (?, ?, ?, ?, ?, NOW())"
	result, err := DbContext.Exec(query, maxID+1, u.Username, u.Email, u.Password, u.Salt)
	if err != nil {
		return fmt.Errorf("erreur lors de la création de l'utilisateur : %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération de l'ID : %v", err)
	}
	u.ID = int(id)
	return nil
}

// GetUserByID récupère un utilisateur par son ID
func GetUserByID(id int) (*User, error) {
	user := &User{}
	query := "SELECT id, username, email, passwd_hash, salt, created_at FROM user WHERE id = ?"
	err := DbContext.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Salt, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("erreur lors de la récupération de l'utilisateur : %v", err)
	}
	return user, nil
}

// GetUserByUsername récupère un utilisateur par son nom d'utilisateur
func GetUserByUsername(username string) (*User, error) {
	user := &User{}
	query := "SELECT id, username, email, passwd_hash, salt, created_at FROM user WHERE username = ?"
	err := DbContext.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Salt, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("erreur lors de la récupération de l'utilisateur : %v", err)
	}
	return user, nil
}

// GetUserByEmail récupère un utilisateur par son email
func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	query := "SELECT id, username, email, passwd_hash, salt, created_at FROM user WHERE email = ?"
	err := DbContext.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Salt, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("erreur lors de la récupération de l'utilisateur : %v", err)
	}
	return user, nil
}

// UpdateUser met à jour les informations d'un utilisateur
func (u *User) UpdateUser() error {
	query := "UPDATE user SET username = ?, email = ?, passwd_hash = ?, salt = ? WHERE id = ?"
	_, err := DbContext.Exec(query, u.Username, u.Email, u.Password, u.Salt, u.ID)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour de l'utilisateur : %v", err)
	}
	return nil
}

// DeleteUser supprime un utilisateur
func (u *User) DeleteUser() error {
	query := "DELETE FROM user WHERE id = ?"
	_, err := DbContext.Exec(query, u.ID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression de l'utilisateur : %v", err)
	}
	return nil
}

func GetUserById(db *sql.DB, id int) (*User, error) {
	row := db.QueryRow("SELECT id, username, email, passwd_hash, salt, created_at FROM user WHERE id = ?", id)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Salt, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func CreateUser(db *sql.DB, username, email, password string) error {
	exists, err := CheckUsernameExists(db, username)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification de l'existence du nom d'utilisateur : %v", err)
	}
	if exists {
		return fmt.Errorf("le nom d'utilisateur %s existe déjà", username)
	}

	maxId, err := utils.GetMaxID("user")
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération de l'ID maximum : %v", err)
	}

	id := maxId + 1

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	saltHex := hex.EncodeToString(salt)

	hash := sha256.Sum256([]byte(password + saltHex))
	hashHex := hex.EncodeToString(hash[:])

	_, err = db.Exec(`
		INSERT INTO user (id, username, email, passwd_hash, salt, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())`,
		id, username, email, hashHex, saltHex)
	return err
}

func CheckUsernameExists(db *sql.DB, username string) (bool, error) {
	row := db.QueryRow("SELECT COUNT(*) FROM user WHERE username = ?", username)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, fmt.Errorf("erreur lors de la vérification de l'existence du nom d'utilisateur : %v", err)
	}
	return count > 0, nil
}

func CheckPassword(db *sql.DB, userID int, password string) (bool, error) {
	user, err := GetUserById(db, userID)
	if err != nil {
		return false, err
	}

	hash := sha256.Sum256([]byte(password + user.Salt))
	hashHex := hex.EncodeToString(hash[:])

	return hashHex == user.Password, nil
}

func ConnectUser(db *sql.DB, username, password string) (*User, error) {
	utils.GetIdBySomething("user", "username", username)
	row := db.QueryRow("SELECT id, username, email, passwd_hash, salt, created_at FROM user WHERE username = ?", username)
	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Salt, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("utilisateur non trouvé")
		}
		return nil, err
	}
	isValid, err := CheckPassword(db, user.ID, password)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la vérification du mot de passe : %v", err)
	}
	if !isValid {
		return nil, fmt.Errorf("mot de passe incorrect")
	}
	return &user, nil
}

func EditPassword(db *sql.DB, userID int, newPassword string) error {
	user, err := GetUserById(db, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération de l'utilisateur : %v", err)
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("erreur lors de la génération du sel : %v", err)
	}
	saltHex := hex.EncodeToString(salt)

	hash := sha256.Sum256([]byte(newPassword + saltHex))
	hashHex := hex.EncodeToString(hash[:])

	_, err = db.Exec(`
		UPDATE user SET passwd_hash = ?, salt = ? WHERE id = ?`,
		hashHex, saltHex, user.ID)
	return err
}

func EditEmail(db *sql.DB, userID int, newEmail string) error {
	_, err := db.Exec("UPDATE user SET email = ? WHERE id = ?", newEmail, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour de l'email : %v", err)
	}
	return nil
}

func EditUsername(db *sql.DB, userID int, newUsername string) error {
	_, err := db.Exec("UPDATE user SET username = ? WHERE id = ?", newUsername, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour du nom d'utilisateur : %v", err)
	}
	return nil
}

// GetInitials retourne les initiales de l'utilisateur
func (u *User) GetInitials() string {
	if len(u.Username) == 0 {
		return "??"
	}
	if len(u.Username) == 1 {
		return strings.ToUpper(u.Username[:1])
	}
	return strings.ToUpper(u.Username[:2])
}
