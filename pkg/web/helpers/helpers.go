package helpers

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, data any) error {
	w.Header().Set("Content-Type", "application/json")

	return json.NewEncoder(w).Encode(data)
}

const (
	ErrCodeInvalidJSON          = "invalid_json"
	ErrCodeInvalidInput         = "invalid_input"
	ErrCodeQuotaExceeded        = "quota_exceeded"
	ErrCodeUnauthorized         = "unauthorized"
	ErrCodeFailedToGenerateUUID = "failed_uuid"
	ErrCodeGetClaims            = "get_claims"
	ErrCodeDBErr                = "db_err"
	ErrCodeInternal             = "internal_error"
	ErrCodeFileNotFound         = "file_not_found"
	ErrCodeForbidden            = "forbidden"
	ErrCodeOffsetMismatch       = "offset_mismatch"
	ErrCodeInvalidRequest       = "invalid_request"
)

type APIErrorResponse struct {
	Error   string `json:"error" example:"Unauthorized"`
	Code    string `json:"code,omitempty" example:"unauthorized"`
	Details string `json:"details,omitempty" example:"invalid or missing join token"`
}

func WriteAPIError(w http.ResponseWriter, errorMsg, code, details string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")

	res := APIErrorResponse{
		Error:   errorMsg,
		Code:    code,
		Details: details,
	}

	dat, marshallErr := json.Marshal(res)
	if marshallErr != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)

	_, writeErr := w.Write(dat)
	if writeErr != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
