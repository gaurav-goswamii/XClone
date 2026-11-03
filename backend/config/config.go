package config

import (
	"os"
	"log"
	"github.com/joho/godotenv"
)

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

func LoadConfig() *Config {
	_ = godotenv.Load(".env")
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
