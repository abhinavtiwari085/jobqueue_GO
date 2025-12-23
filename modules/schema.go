package modules

// Job represents a row in jobs table
type Job struct {
	JobID        int64  `json:"job_id"`
	Command      string `json:"command"`
	State        string `json:"state"`
	AttemptCount int    `json:"attempt_counts"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// Job states
const (
	STATE_WAITING     = "WAITING"
	STATE_IN_PROGRESS = "IN_PROGRESS"
	STATE_COMPLETED   = "COMPLETED"
	STATE_FAILED      = "FAILED"
)

// Job table schema
const JobSchema = `
CREATE TABLE IF NOT EXISTS jobs (
	job_id BIGSERIAL PRIMARY KEY,
	command TEXT NOT NULL,
	state VARCHAR(20) NOT NULL,
	attempt_counts INT NOT NULL DEFAULT 0,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`
