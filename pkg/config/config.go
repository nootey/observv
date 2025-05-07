package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type Config struct {
	DatabaseName string
	Host         string
	HttpPort     string
	Release      bool
}

func LoadConfig() *Config {

	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	release, err := strconv.ParseBool(os.Getenv("RELEASE"))
	if err != nil {
		log.Fatalf("Invalid release mode: %v", err)
	}

	host := os.Getenv("HTTP_HOST")
	if host == "" {
		log.Println("Host not defined in environment variables, using global fallback ...")
		host = "0.0.0.0"
	}

	return &Config{
		DatabaseName: os.Getenv("DATABASE_NAME"),
		Host:         os.Getenv("HOST"),
		HttpPort:     os.Getenv("HTTP_PORT"),
		Release:      release,
	}
}
