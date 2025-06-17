package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	Host       string
	Port       string
	SessionKey string
}

var AppConfig Config

// LoadConfig charge les variables d'environnement depuis le fichier .env
func LoadConfig() error {
	err := godotenv.Load("forum.env")
	if err != nil {
		return fmt.Errorf("erreur lors du chargement du fichier forum.env: %v", err)
	}

	AppConfig = Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		Host:       os.Getenv("HOST"),
		Port:       os.Getenv("PORT"),
		SessionKey: os.Getenv("SESSION_KEY"),
	}

	// Vérification des variables obligatoires
	if AppConfig.DBHost == "" || AppConfig.DBPort == "" || AppConfig.DBUser == "" ||
		AppConfig.DBName == "" || AppConfig.Host == "" || AppConfig.Port == "" ||
		AppConfig.SessionKey == "" {
		return fmt.Errorf("toutes les variables d'environnement requises doivent être définies")
	}

	return nil
}

// GetDSN retourne la chaîne de connexion à la base de données
func GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		AppConfig.DBUser,
		AppConfig.DBPassword,
		AppConfig.DBHost,
		AppConfig.DBPort,
		AppConfig.DBName)
}
