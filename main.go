package main

import (
	"database/sql"
	"fmt"
	"jobqueue/config"
	"jobqueue/modules"
	"jobqueue/services"
	"log"
	"os/exec"
	"sync"

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
		fmt.Printf("Running job %s (attempt %d)\n", job.JobID, attempt)

		cmd := exec.Command("cmd", "/C", job.Command)
		output, err := cmd.CombinedOutput()

		if err == nil {
			fmt.Printf("Job %s succeeded\n", job.JobID)
			fmt.Println("Output:", string(output))
			_ = services.UpdateJobState(db, job.JobID, "COMPLETED", false)
			return
		}

		fmt.Printf("Job %s failed (attempt %d): %v\n", job.JobID, attempt, err)
		fmt.Println("Output:", string(output))

		// increment attempt count in DB and keep as in-progress
		_ = services.UpdateJobState(db, job.JobID, "IN_PROGRESS", true)

		if attempt == maxAttempts {
			fmt.Printf("Job %s reached max attempts\n", job.JobID)
			_ = services.UpdateJobState(db, job.JobID, "FAILED", false)
			return
		}
	}
}

func worker(workerID int, db *sql.DB, jobQueue <-chan *modules.Job, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobQueue {
		fmt.Printf("[worker %d] picked job %s\n", workerID, job.JobID)
		runJobLogic(db, job)
	}
}

func main() {
	// Load config and connect DB
	config.LoadEnv()
	db, err := config.DbConfig()
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		log.Fatal(err)
	}
	defer db.Close()

	// Create jobs (same as your example)
	createdJobID1, err := services.CreateJob(db, "ping google.com")
	if err != nil {
		log.Fatal("Error creating job:", err)
	}
	fmt.Println("Created job with ID:", createdJobID1)

	createdJobID2, err := services.CreateJob(db, "echo hello world")
	if err != nil {
		log.Fatal("Error creating job:", err)
	}
	fmt.Println("Created job with ID:", createdJobID2)

	createdJobID3, err := services.CreateJob(db, "ech hello world") // will fail
	if err != nil {
		log.Fatal("Error creating job:", err)
	}
	fmt.Println("Created job with ID:", createdJobID3)

	// Fetch waiting jobs
	waitingJobs, err := services.GetWaitingJobs(db)
	if err != nil {
		fmt.Println("Error fetching waiting jobs:", err)
		log.Fatal(err)
	}
	fmt.Println("Waiting Jobs:", len(waitingJobs))

	// --- Worker pool setup ---
	const workerCount = 3
	jobQueueBufferSize := len(waitingJobs) // buffer enough for all jobs (or use a fixed number like 100)

	jobQueue := make(chan *modules.Job, jobQueueBufferSize)
	var workersWg sync.WaitGroup

	// Start N workers
	for workerID := 1; workerID <= workerCount; workerID++ {
		workersWg.Add(1)
		go worker(workerID, db, jobQueue, &workersWg)
	}

	// Send jobs to workers
	for _, job := range waitingJobs {
		jobQueue <- job
	}

	// Close queue so workers exit after finishing
	close(jobQueue)

	// Wait for all workers to finish
	workersWg.Wait()

	fmt.Println("All jobs completed (workers stopped)")
}
