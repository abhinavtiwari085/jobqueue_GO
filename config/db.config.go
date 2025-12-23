package config

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/lib/pq"
)

// CREATE TABLE jobs(job_id BIGSERIAL PRIMARY KEY, command TEXT NOT NULL, state VARCHAR(20) NOT NULL, attempt_counts INT NOT NULL DEFAULT 0);
func DbConfig() (*sql.DB, error) {
	port, err := strconv.Atoi(DB_PORT)
	if err != nil {
		panic("Invalid DB_PORT")
	}

	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		DB_HOST, port, DB_USER, DB_PASSWORD, DB_NAME,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	return db, nil

}
