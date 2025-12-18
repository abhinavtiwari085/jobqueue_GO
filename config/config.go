package config

import (
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	DB_HOST     string
	DB_PORT     string
	DB_USER     string
	DB_PASSWORD string
	DB_NAME     string
)

func LoadEnv() {
	DB_HOST = os.Getenv("DB_HOST")
	if DB_HOST == "" {
		fmt.Println("Error: DB_HOST environment variable is not set")
	}

	DB_PORT = os.Getenv("DB_PORT")
	if DB_PORT == "" {
		fmt.Println("Error: DB_PORT environment variable is not set")
	}

	DB_USER = os.Getenv("DB_USER")
	if DB_USER == "" {
		fmt.Println("Error: DB_USER environment variable is not set")
	}

	DB_PASSWORD = os.Getenv("DB_PASSWORD")
	if DB_PASSWORD == "" {
		fmt.Println("Error: DB_PASSWORD environment variable is not set")
	}

	DB_NAME = os.Getenv("DB_NAME")
	if DB_NAME == "" {
		fmt.Println("Error: DB_NAME environment variable is not set")
	}
}
