package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type Config struct {
	Host             string
	HttpPort         string
	Release          bool
	DBHost           string
	DBUser           string
	DBPassword       string
	DBPort           int
	Database         string
	RootUserPassword string
	AppVersion       string
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

	DBPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatalf("Invalid DB port: %v", err)
	}

	return &Config{
		Host:             os.Getenv("HTTP_HOST"),
		HttpPort:         os.Getenv("HTTP_PORT"),
		DBHost:           os.Getenv("DB_HOST"),
		DBUser:           os.Getenv("DB_USER"),
		DBPassword:       os.Getenv("DB_PASSWORD"),
		DBPort:           DBPort,
		Database:         os.Getenv("DATABASE"),
		Release:          release,
		RootUserPassword: os.Getenv("ROOT_USER_PASSWORD"),
	}
}
