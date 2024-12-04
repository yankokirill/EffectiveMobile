package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

var config Config

type Config struct {
	serverAddress  string
	databaseURL    string
	externalApiURL string
}

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("WARNING: No .env file found")
	}

	config = Config{
		serverAddress:  os.Getenv("SERVER_ADDRESS"),
		databaseURL:    os.Getenv("DATABASE_URL"),
		externalApiURL: os.Getenv("EXTERNAL_API_URL"),
	}

	if config.serverAddress == "" {
		config.serverAddress = ":8080"
		log.Println("WARNING: SERVER_ADDRESS environment variable not set")
	}
	if config.databaseURL == "" {
		config.databaseURL = "postgres://user:password@localhost:5432/db"
		log.Println("WARNING: DATABASE_URL environment variable not set")
	}
	if config.externalApiURL == "" {
		config.externalApiURL = "http://localhost:8081"
		log.Println("WARNING: EXTERNAL_API_URL environment variable not set")
	}
}

func ServerAddress() string {
	return config.serverAddress
}

func DatabaseURL() string {
	return config.databaseURL
}

func ExternalApiURL() string {
	return config.externalApiURL
}
