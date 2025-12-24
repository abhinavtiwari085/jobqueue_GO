package main

import (
	"fmt"
	"jobqueue/config"
	"jobqueue/routes"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

func main() {
	// Load config and connect DB
	config.LoadEnv()
	db, err := config.DbConfig()
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		log.Fatal(err)
	}
	defer db.Close()

	if len(os.Args) < 2 {
		fmt.Println("wrong command")
		routes.Help()
		return
	}

	if os.Args[1] == "jobctl" {
		routes.Dispatch(db, os.Args[2:])
		return
	} else {
		fmt.Println("wrong command")
		routes.Help()
		return
	}

}
