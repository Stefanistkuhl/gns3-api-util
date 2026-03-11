package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/messageUtils"
	"github.com/0xveya/gns3util/pkg/models"
)

type APIError struct {
	StatusCode int
	ErrorMsg   string `json:"error"`
	Code       string `json:"code"`
	Details    string `json:"details"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("api error %d: %s (code: %s, details: %s)", e.StatusCode, e.ErrorMsg, e.Code, e.Details)
}

type ClientV2 struct {
	base *BaseClient
}

func NewClientV2(settings Settings) *ClientV2 {
	return &ClientV2{
		base: NewBaseClient(settings),
	}
}

func (c *ClientV2) Do(ctx context.Context, opts *requestOptions) ([]byte, *http.Response, error) {
	body, resp, err := c.base.Do(ctx, opts)
	if err != nil {
		return nil, nil, err
	}

	if opts.stream {
		return nil, resp, nil
	}

	if resp.StatusCode >= 300 {
		return body, resp, c.parseAPIError(body, resp.StatusCode)
	}

	return body, resp, nil
}

func (c *ClientV2) parseAPIError(body []byte, statusCode int) error {
	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err == nil {
		apiErr.StatusCode = statusCode
		if apiErr.ErrorMsg != "" || apiErr.Code != "" || apiErr.Details != "" {
			return &apiErr
		}
	}

	return fmt.Errorf("%s", messageUtils.WarningMsgf(
		"unexpected response (status %d): %s",
		statusCode, string(body),
	))
}

func (c *ClientV2) GetAuthStatus(ctx context.Context) (*models.AuthStatusResponse, error) {
	opts := NewRequestOptions(c.base.settings).
		WithURL("/auth/status").
		WithMethod(GET)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.AuthStatusResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resp, nil
}
