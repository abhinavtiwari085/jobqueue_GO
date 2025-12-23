package services

import (
	"database/sql"
	"fmt"
	"jobqueue/modules"
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
//fetch job by id in waiting state
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

