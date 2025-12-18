package main

import (
	"fmt"
	"jobqueue/config"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("Hello, World!")
	config.LoadEnv()
	config.DbConfig()
}
