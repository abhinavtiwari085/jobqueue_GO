package controllers

import (
	"database/sql"
	"fmt"
	"jobqueue/modules"
	"jobqueue/services"
	"log"
	"sync"
)

var jobQueue chan *modules.Job
var workersWg sync.WaitGroup
var workerCount int
var workerRunning bool
var workerMu sync.Mutex

func worker(workerID int, db *sql.DB, jobQueue <-chan *modules.Job, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobQueue {
		fmt.Printf("[worker %d] picked job %d\n", workerID, job.JobID)
		services.RunJobLogic(db, job)
	}
}

func StartWorkerController(db *sql.DB, cnt int) {
	workerMu.Lock()
	if workerRunning {
		workerMu.Unlock()
		fmt.Println("Workers already running")
		return
	}
	workerRunning = true
	workerMu.Unlock()

	waitingJobs, err := services.GetWaitingJobs(db)
	if err != nil {
		fmt.Println("Error fetching waiting jobs:", err)
		log.Fatal(err)
	}
	fmt.Println("Waiting Jobs:", len(waitingJobs))

	workerCount = cnt
	jobQueueBufferSize := len(waitingJobs)
	if jobQueueBufferSize == 0 {
		jobQueueBufferSize = 1
	}
	jobQueue = make(chan *modules.Job, jobQueueBufferSize)

	for workerID := 1; workerID <= workerCount; workerID++ {
		workersWg.Add(1)
		go worker(workerID, db, jobQueue, &workersWg)
	}

	for _, job := range waitingJobs {
		jobQueue <- job
	}
}

func StopWorkerController(db *sql.DB) {
	workerMu.Lock()
	if !workerRunning {
		workerMu.Unlock()
		fmt.Println("Workers are not running")
		return
	}
	workerRunning = false
	workerMu.Unlock()

	close(jobQueue)
	workersWg.Wait()
}

func CreateJobController(db *sql.DB, args []string) {
	if len(args > 0) {
		createdJobID1, err := services.CreateJob(db, args[0])
		if err != nil {
			log.Fatal("Error creating job:", err)
		}
		fmt.Println("Created job with ID:", createdJobID1)
	} else {
		fmt.Println("provide command to create job")
	}
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
