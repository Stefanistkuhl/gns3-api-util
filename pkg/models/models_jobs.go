package models

import "time"

type JobInfo struct {
	Name         string `json:"name"`
	NodeID       string `json:"node_id"`
	Interval     string `json:"interval"`
	Description  string `json:"description"`
	RegisteredAt string `json:"registered_at"`
}

type ListJobsResponse struct {
	Jobs  []JobInfo `json:"jobs"`
	Count int       `json:"count"`
}

type RunJobRequest struct {
	NodeID string `json:"node_id"`
}

type RunJobResponse struct {
	RunID   string `json:"run_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type JobRunStatus struct {
	RunID       string     `json:"run_id"`
	JobName     string     `json:"job_name"`
	NodeID      string     `json:"node_id"`
	Status      string     `json:"status"`
	InvokedBy   string     `json:"invoked_by"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Result      any        `json:"result,omitempty"`
}

type ListJobRunsResponse struct {
	Runs  []JobRunStatus `json:"runs"`
	Count int            `json:"count"`
}
