package main

import (
	"database/sql"
	"fmt"
	"jobqueue/config"
	"jobqueue/modules"
	"jobqueue/services"
	"log"
	"os/exec"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

func runJobLogic(db *sql.DB, job *modules.Job) {
	const maxAttempts = 3

	if job.AttemptCount >= maxAttempts {
		fmt.Printf("Job %s already exceeded max attempts\n", job.JobID)
		_ = services.UpdateJobState(db, job.JobID, "FAILED", false)
		return
	}

	for attempt := job.AttemptCount + 1; attempt <= maxAttempts; attempt++ {
		fmt.Printf("Running job %d (attempt %d)\n", job.JobID, attempt)

		cmd := exec.Command("cmd", "/C", job.Command)
		output, err := cmd.CombinedOutput()

		if err == nil {
			fmt.Printf("Job %d succeeded\n", job.JobID)
			fmt.Println("Output:", string(output))
			_ = services.UpdateJobState(db, job.JobID, "COMPLETED", false)
			break
		}

		fmt.Printf("Job %s failed: %v\n", job.JobID, err)
		fmt.Println("Output:", string(output))

		_ = services.UpdateJobState(db, job.JobID, "IN_PROGRESS", true)

		if attempt == maxAttempts {
			fmt.Printf("Job %d reached max attempts\n", job.JobID)
			_ = services.UpdateJobState(db, job.JobID, "FAILED", false)
		}
	}
}

func main() {
	// Get DB connection
	config.LoadEnv()
	db, err := config.DbConfig()
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		log.Fatal(err)
	}
	defer db.Close()

	// Call CreateJob
	jobID, err := services.CreateJob(db, "ping google.com")
	if err != nil {
		fmt.Println("Error creating job:", err)
		log.Fatal(err)
	}

	fmt.Println("Created job with ID:", jobID)

	jobID2, err := services.CreateJob(db, "echo hello world")
	if err != nil {
		fmt.Println("Error creating job:", err)
		log.Fatal(err)
	}
	fmt.Println("Created job with ID:", jobID2)

	jobID3, err := services.CreateJob(db, "ech hello world")
	if err != nil {
		fmt.Println("Error creating job:", err)
		log.Fatal(err)
	}

	fmt.Println("Created job with ID:", jobID3)

	// Fetch waiting jobs
	jobs, err := services.GetWaitingJobs(db)
	if err != nil {
		fmt.Println("Error fetching waiting jobs:", err)
		log.Fatal(err)
	}
	fmt.Println("Waiting Jobs:")

	for _, job := range jobs {
		//logic area start
		runJobLogic(db, job)
		//logic area end
	}
}
