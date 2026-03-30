package handlers

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/state/pb"
	"github.com/0xveya/gns3util/pkg/web/helpers"
	"github.com/0xveya/gns3util/pkg/web/middleware"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListJobs returns all registered jobs from etcd
//
//	@Summary		List registered jobs
//	@Description	Returns all background jobs registered by cluster nodes
//	@Tags			jobs
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	models.ListJobsResponse
//	@Failure		500	{object}	helpers.APIErrorResponse
//	@Router			/api/v1/jobs [get]
func (m *Master) ListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := m.Store.GetRegisteredJobs(r.Context())
	if err != nil {
		helpers.WriteAPIError(w, "Failed to list jobs", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	var jobInfos []models.JobInfo
	for _, job := range jobs {
		registeredAt := ""
		if job.RegisteredAt != nil {
			registeredAt = job.RegisteredAt.AsTime().Format(time.RFC3339)
		}
		jobInfos = append(jobInfos, models.JobInfo{
			Name:         job.Name,
			NodeID:       job.NodeId,
			Interval:     job.Interval,
			Description:  job.Description,
			RegisteredAt: registeredAt,
		})
	}

	resp := models.ListJobsResponse{
		Jobs:  jobInfos,
		Count: len(jobInfos),
	}

	if writeErr := helpers.WriteJSON(w, resp); writeErr != nil {
		m.Logger.Error("Failed to write list jobs response", "err", writeErr)
	}
}

// RunJob triggers a job execution on a specific node
//
//	@Summary		Run a job
//	@Description	Triggers a background job on the target node, returns a run ID for tracking
//	@Tags			jobs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			job_name	path		string					true	"Job name"
//	@Param			request		body		models.RunJobRequest	true	"Run job request"
//	@Success		202			{object}	models.RunJobResponse
//	@Failure		400			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/jobs/{job_name}/run [post]
func (m *Master) RunJob(w http.ResponseWriter, r *http.Request) {
	jobName := chi.URLParam(r, "job_name")

	var req models.RunJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteAPIError(w, "Invalid request body", helpers.ErrCodeInvalidRequest, err.Error(), http.StatusBadRequest)
		return
	}

	if req.NodeID == "" {
		helpers.WriteAPIError(w, "node_id is required", helpers.ErrCodeInvalidInput, "node_id must be specified", http.StatusBadRequest)
		return
	}

	nodes, err := m.Store.GetNodes(r.Context())
	if err != nil {
		helpers.WriteAPIError(w, "Failed to get nodes", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	var targetNode *pb.Node
	for _, n := range nodes {
		if n.Id == req.NodeID {
			targetNode = n
			break
		}
	}

	if targetNode == nil {
		helpers.WriteAPIError(w, "Node not found", helpers.ErrCodeNotFound, "node "+req.NodeID+" not found", http.StatusNotFound)
		return
	}

	runUUID, uuidErr := uuid.NewV7()
	if uuidErr != nil {
		helpers.WriteAPIError(w, "Failed to generate run ID", helpers.ErrCodeInternal, uuidErr.Error(), http.StatusInternalServerError)
		return
	}
	runID := runUUID.String()

	claims, ok := middleware.GetClaims(r)
	if !ok {
		helpers.WriteAPIError(w, "Failed to get claims", helpers.ErrCodeGetClaims, "missing claims", http.StatusInternalServerError)
		return
	}
	authToken := r.Header.Get("Authorization")

	jobRun := &pb.JobRun{
		RunId:     runID,
		JobName:   jobName,
		NodeId:    req.NodeID,
		Status:    "running",
		InvokedBy: claims.UserID,
		StartedAt: timestamppb.Now(),
	}

	if err := m.Store.PutJobRun(r.Context(), jobRun); err != nil {
		helpers.WriteAPIError(w, "Failed to store job run", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	go m.executeJobOnNode(targetNode, jobName, runID, authToken)

	resp := models.RunJobResponse{
		RunID:   runID,
		Status:  "running",
		Message: fmt.Sprintf("Job %s triggered on node %s", jobName, req.NodeID),
	}

	w.WriteHeader(http.StatusAccepted)
	if writeErr := helpers.WriteJSON(w, resp); writeErr != nil {
		m.Logger.Error("Failed to write run job response", "err", writeErr)
	}
}

func (m *Master) executeJobOnNode(node *pb.Node, jobName, runID, authToken string) {
	nodeURL := fmt.Sprintf("https://%s:%d/api/v1/jobs/%s/run", node.IpAddress, node.ApiPort, jobName)

	caPath := filepath.Join(filepath.Clean(m.TLSDir), "ca.crt")
	caData, err := os.ReadFile(caPath) //#nosec G304
	if err != nil {
		m.Logger.Error("Failed to read CA cert for job execution", "err", err)
		m.updateJobRunFailed(runID, "failed to read CA cert: "+err.Error())
		return
	}

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caData)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caPool,
			},
		},
		Timeout: 5 * time.Minute,
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, nodeURL, strings.NewReader("{}"))
	if err != nil {
		m.Logger.Error("Failed to create job request", "err", err)
		m.updateJobRunFailed(runID, "failed to create request: "+err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		m.Logger.Error("Failed to execute job on node", "err", err, "node_id", node.Id)
		m.updateJobRunFailed(runID, "request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		m.Logger.Error("Job execution returned error", "status", resp.StatusCode, "body", string(body))
		m.updateJobRunFailed(runID, fmt.Sprintf("node returned status %d: %s", resp.StatusCode, string(body)))
		return
	}

	m.updateJobRunCompleted(runID, body)
}

func (m *Master) updateJobRunFailed(runID, errMsg string) {
	run, err := m.Store.GetJobRun(context.Background(), runID)
	if err != nil {
		m.Logger.Error("Failed to get job run for update", "run_id", runID, "err", err)
		return
	}

	run.Status = "failed"
	run.CompletedAt = timestamppb.Now()
	run.Result = []byte(fmt.Sprintf(`{"error": %q}`, errMsg))

	if err := m.Store.PutJobRun(context.Background(), run); err != nil {
		m.Logger.Error("Failed to update failed job run", "run_id", runID, "err", err)
	}
}

func (m *Master) updateJobRunCompleted(runID string, result []byte) {
	run, err := m.Store.GetJobRun(context.Background(), runID)
	if err != nil {
		m.Logger.Error("Failed to get job run for update", "run_id", runID, "err", err)
		return
	}

	run.Status = "completed"
	run.CompletedAt = timestamppb.Now()
	run.Result = result

	if err := m.Store.PutJobRun(context.Background(), run); err != nil {
		m.Logger.Error("Failed to update completed job run", "run_id", runID, "err", err)
	}
}

// GetJobRun returns the status and result of a specific job run
//
//	@Summary		Get job run result
//	@Description	Returns the status and result of a specific job run by its UUID
//	@Tags			jobs
//	@Produce		json
//	@Security		BearerAuth
//	@Param			run_id	path		string	true	"Job run UUID"
//	@Success		200		{object}	models.JobRunStatus
//	@Failure		404		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/jobs/runs/{run_id} [get]
func (m *Master) GetJobRun(w http.ResponseWriter, r *http.Request) {
	runID := chi.URLParam(r, "run_id")

	run, err := m.Store.GetJobRun(r.Context(), runID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			helpers.WriteAPIError(w, "Job run not found", helpers.ErrCodeNotFound, err.Error(), http.StatusNotFound)
			return
		}
		helpers.WriteAPIError(w, "Failed to get job run", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	status := models.JobRunStatus{
		RunID:     run.RunId,
		JobName:   run.JobName,
		NodeID:    run.NodeId,
		Status:    run.Status,
		InvokedBy: run.InvokedBy,
		StartedAt: run.StartedAt.AsTime(),
	}

	if run.CompletedAt != nil {
		t := run.CompletedAt.AsTime()
		status.CompletedAt = &t
	}

	if len(run.Result) > 0 {
		var result any
		if err := json.Unmarshal(run.Result, &result); err == nil {
			status.Result = result
		} else {
			status.Result = string(run.Result)
		}
	}

	if writeErr := helpers.WriteJSON(w, status); writeErr != nil {
		m.Logger.Error("Failed to write job run response", "err", writeErr)
	}
}

// ListJobRuns returns all job runs
//
//	@Summary		List job runs
//	@Description	Returns all job run entries
//	@Tags			jobs
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	models.ListJobRunsResponse
//	@Failure		500	{object}	helpers.APIErrorResponse
//	@Router			/api/v1/jobs/runs [get]
func (m *Master) ListJobRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := m.Store.ListJobRuns(r.Context())
	if err != nil {
		helpers.WriteAPIError(w, "Failed to list job runs", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	var statuses []models.JobRunStatus
	for _, run := range runs {
		status := models.JobRunStatus{
			RunID:     run.RunId,
			JobName:   run.JobName,
			NodeID:    run.NodeId,
			Status:    run.Status,
			InvokedBy: run.InvokedBy,
			StartedAt: run.StartedAt.AsTime(),
		}

		if run.CompletedAt != nil {
			t := run.CompletedAt.AsTime()
			status.CompletedAt = &t
		}

		statuses = append(statuses, status)
	}

	resp := models.ListJobRunsResponse{
		Runs:  statuses,
		Count: len(statuses),
	}

	if writeErr := helpers.WriteJSON(w, resp); writeErr != nil {
		m.Logger.Error("Failed to write list job runs response", "err", writeErr)
	}
}
