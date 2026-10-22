package config

import (
	"os"
	"log"
	"github.com/joho/godotenv"
)

// everything backend requires
type Config struct {
	Port                   string
	Environment            string
	DatabaseURL            string
	JWTSecret              string
	CloudinaryCloudName    string
	CloudinaryAPIKey       string
	CloudinaryAPISecret    string
	CloudinaryUploadFolder string
}

// LoadConfig loads configuration from environment variables or defaults
func LoadConfig() *Config {
	// Load env file from /backend dir
	_ = godotenv.Load(".env")
	// collecting config values
	port := os.Getenv("PORT")
	env := os.Getenv("ENV")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}  
	cloudinaryCloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	cloudinaryAPIKey := os.Getenv("CLOUDINARY_API_KEY")
	cloudinaryAPISecret := os.Getenv("CLOUDINARY_API_SECRET")
	cloudinaryUploadFolder := os.Getenv("CLOUDINARY_UPLOAD_FOLDER")
	// A warning, not log.Fatal: unlike DATABASE_URL and JWT_SECRET, photos are
	// optional, and failing hard here would stop anyone without a Cloudinary
	// account from running the app at all. The upload and create endpoints
	// answer 503 at request time, so this makes that outcome visible at boot.
	if cloudinaryCloudName == "" || cloudinaryAPIKey == "" || cloudinaryAPISecret == "" {
		log.Printf("⚠️  CLOUDINARY_* not fully configured: avatar and post photo uploads will return 503")
	}
	return &Config{
		Port:                   port,
		Environment:            env,
		DatabaseURL:            dbURL,
		JWTSecret:              jwtSecret,
		CloudinaryCloudName:    cloudinaryCloudName,
		CloudinaryAPIKey:       cloudinaryAPIKey,
		CloudinaryAPISecret:    cloudinaryAPISecret,
		CloudinaryUploadFolder: cloudinaryUploadFolder,
	}
}
