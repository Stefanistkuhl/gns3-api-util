package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	backgroundjobs "github.com/0xveya/gns3util/internal/file-store/backroundjobs"
	"github.com/0xveya/gns3util/pkg/web/helpers"
	"github.com/go-chi/chi/v5"
)

type JobRunnerFunc func(ctx context.Context, invokedBy backgroundjobs.Invocator) (any, error)

// RunJob executes a background job
//
//	@Summary		Run background job
//	@Description	Triggers execution of a registered background job
//	@Tags			jobs
//	@Produce		json
//	@Param			job_name	path		string	true	"Job name to execute"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/jobs/{job_name} [post]
//	@Security		BearerAuth
func (h *FilestoreHandlers) RunJob(w http.ResponseWriter, r *http.Request) {
	jobName := chi.URLParam(r, "job_name")

	runner, ok := h.JobRunners[jobName]
	if !ok {
		helpers.WriteAPIError(w, "Job not found", helpers.ErrCodeNotFound, "no such job: "+jobName, http.StatusNotFound)
		return
	}

	result, err := runner(r.Context(), backgroundjobs.InvocatorMaster)
	if err != nil {
		helpers.WriteAPIError(w, "Job execution failed", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(result); encodeErr != nil {
		helpers.WriteAPIError(w, "Failed to encode response", helpers.ErrCodeInternal, encodeErr.Error(), http.StatusInternalServerError)
	}
}
