package controllers

import (
	"database/sql"
	"fmt"
	"jobqueue/modules"
	"jobqueue/services"
	"log"
	"sync"
)

func worker(workerID int, db *sql.DB, jobQueue <-chan *modules.Job, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobQueue {
		fmt.Printf("[worker %d] picked job %s\n", workerID, job.JobID)
		services.RunJobLogic(db, job)
	}
}

func CreateJobController(db *sql.DB, args []string) {
	createdJobID1, err := services.CreateJob(db, args[0])
	if err != nil {
		log.Fatal("Error creating job:", err)
	}
	fmt.Println("Created job with ID:", createdJobID1)
}

func StartWorkerController(db *sql.DB) {
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
}

func StopWorkerController() {

}

func ListJobsController(db *sql.DB, args []string) {
	jobs, err := services.GetJobsByState(db, args[0])
	if err != nil {
		fmt.Println("Error fetching  jobs of state:", err)
		log.Fatal(err)
	}
	fmt.Println(" Jobs of state:", args[0], "=", len(jobs))

	for _, job := range jobs {
		fmt.Println("id: ", job.JobID, " command: ", job.Command, " state:", job.State, " attempt count:", job.AttemptCount, " created at:", job.CreatedAt, " updated at:", job.UpdatedAt)
	}

}

func GetStatusController(db *sql.DB) {
	counts, err := services.GetJobCountByState(db)
	if err != nil {
		fmt.Println("Error getting job counts by state:", err)
		log.Fatal(err)
	}
	fmt.Println("Job counts by state:", counts)
}
