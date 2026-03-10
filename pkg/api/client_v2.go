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

func (c *ClientV2) Do(opts *requestOptions) ([]byte, *http.Response, error) {
	body, resp, err := c.base.DOv2(opts)
	if err != nil {
		if resp == nil {
			return nil, nil, err
		}

		if resp.StatusCode >= 300 {
			var apiErr APIError
			if parseErr := json.Unmarshal(body, &apiErr); parseErr == nil {
				apiErr.StatusCode = resp.StatusCode

				if apiErr.ErrorMsg == "" && apiErr.Code == "" && apiErr.Details == "" {
					hint := messageUtils.WarningMsgf(
						"received unexpected response format from server. "+
							"This usually indicates the wrong server URL. "+
							"Raw response: %s",
						string(body),
					)
					return body, resp, fmt.Errorf("%s", hint)
				}

				return body, resp, &apiErr
			}

			hint := messageUtils.WarningMsgf(
				"failed to parse API error response. "+
					"Possible wrong server URL. (status %d): %s",
				resp.StatusCode, string(body),
			)
			return body, resp, fmt.Errorf("%s", hint)
		}
	}

	return body, resp, nil
}

func (c *ClientV2) GetAuthStatus(ctx context.Context) (*models.AuthStatusResponse, error) {
	opts := NewRequestOptions(c.base.settings).
		WithURL("/auth/status").
		WithMethod(GET)

	body, _, err := c.Do(opts)
	if err != nil {
		return nil, err
	}

	var resp models.AuthStatusResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resp, nil
}
