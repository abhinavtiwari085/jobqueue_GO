package services

import (
	"database/sql"
	"fmt"
	"jobqueue/modules"
	"os/exec"
)

func CreateJob(db *sql.DB, command string) (int64, error) {
	var jobID int64

	err := db.QueryRow(
		`INSERT INTO jobs (command, state, attempt_counts)
		 VALUES ($1, $2, $3)
		 RETURNING job_id`,
		command,
		modules.STATE_WAITING,
		0,
	).Scan(&jobID)

	if err != nil {
		return 0, fmt.Errorf("failed to create job: %w", err)
	}

	fmt.Println("Job created with ID:", jobID)
	return jobID, nil
}

// fetch job by id in waiting state
func GetJobByID(db *sql.DB, jobID int64) (*modules.Job, error) {
	var job modules.Job
	err := db.QueryRow(
		`SELECT job_id, command, state, attempt_counts, created_at, updated_at
		 FROM jobs
		 WHERE job_id = $1`,
		jobID,
	).Scan(
		&job.JobID,
		&job.Command,
		&job.State,
		&job.AttemptCount,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No job found
		}
		return nil, fmt.Errorf("failed to fetch job by ID: %w", err)
	}

	return &job, nil
}

// fetch job by id, update job state, increment attempt count, etc. can be added here
func GetWaitingJobs(db *sql.DB) ([]*modules.Job, error) {
	rows, err := db.Query(
		`SELECT job_id, command, state, attempt_counts, created_at, updated_at
		 FROM jobs
		 WHERE state = $1
		 ORDER BY created_at ASC`,
		"WAITING",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query waiting jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*modules.Job

	for rows.Next() {
		var job modules.Job
		err := rows.Scan(
			&job.JobID,
			&job.Command,
			&job.State,
			&job.AttemptCount,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		jobs = append(jobs, &job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return jobs, nil
}

// update jobstate, increment attempt count
func UpdateJobState(db *sql.DB, jobID int64, newState string, incrementAttempt bool) error {
	if incrementAttempt {
		_, err := db.Exec(`
            UPDATE jobs
            SET state = $1,
                attempt_counts = attempt_counts + 1,
                updated_at = NOW()
            WHERE job_id = $2
        `, newState, jobID)
		return err
	}

	_, err := db.Exec(`
        UPDATE jobs
        SET state = $1,
            updated_at = NOW()
        WHERE job_id = $2
    `, newState, jobID)
	return err
}

func RunJobLogic(db *sql.DB, job *modules.Job) {
	const maxAttempts = 3

	if job.AttemptCount >= maxAttempts {
		fmt.Printf("Job %d already exceeded max attempts\n", job.JobID)
		_ = UpdateJobState(db, job.JobID, "FAILED", false)
		return
	}

	for attempt := job.AttemptCount + 1; attempt <= maxAttempts; attempt++ {
		fmt.Printf("Running job %d (attempt %d)\n", job.JobID, attempt)

		cmd := exec.Command("cmd", "/C", job.Command)
		output, err := cmd.CombinedOutput()

		if err == nil {
			fmt.Printf("Job %d succeeded\n", job.JobID)
			fmt.Println("Output:", string(output))
			_ = UpdateJobState(db, job.JobID, "COMPLETED", false)
			return
		}

		fmt.Printf("Job %d failed (attempt %d): %v\n", job.JobID, attempt, err)
		fmt.Println("Output:", string(output))

		// increment attempt count in DB and keep as in-progress
		_ = UpdateJobState(db, job.JobID, "IN_PROGRESS", true)

		if attempt == maxAttempts {
			fmt.Printf("Job %d reached max attempts\n", job.JobID)
			_ = UpdateJobState(db, job.JobID, "FAILED", false)
			return
		}
	}
}

func GetJobsByState(db *sql.DB, state string) ([]*modules.Job, error) {
	rows, err := db.Query(
		`SELECT job_id, command, state, attempt_counts, created_at, updated_at
		 FROM jobs
		 WHERE state = $1
		 ORDER BY created_at ASC`,
		state,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs by state: %w", err)
	}
	defer rows.Close()

	var jobs []*modules.Job

	for rows.Next() {
		var job modules.Job
		err := rows.Scan(
			&job.JobID,
			&job.Command,
			&job.State,
			&job.AttemptCount,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		jobs = append(jobs, &job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return jobs, nil
}


func GetJobCountByState(db *sql.DB) (map[string]int, error) {
	rows, err := db.Query(`
		SELECT state, COUNT(*)
		FROM jobs
		GROUP BY state
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get job counts: %w", err)
	}
	defer rows.Close()

	result := make(map[string]int)

	for rows.Next() {
		var state string
		var count int
		if err := rows.Scan(&state, &count); err != nil {
			return nil, fmt.Errorf("failed to scan count: %w", err)
		}
		result[state] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return result, nil
}
