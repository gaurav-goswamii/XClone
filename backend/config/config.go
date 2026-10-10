package config

import (
	"os"

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
	jwtSecret := os.Getenv("JWT_SECRET")
	cloudinaryCloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	cloudinaryAPIKey := os.Getenv("CLOUDINARY_API_KEY")
	cloudinaryAPISecret := os.Getenv("CLOUDINARY_API_SECRET")
	cloudinaryUploadFolder := os.Getenv("CLOUDINARY_UPLOAD_FOLDER")
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
