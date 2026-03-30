package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

func NewClientV2(settings *Settings) *ClientV2 {
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
		statusCode, strings.TrimSpace(string(body)),
	))
}

func (c *ClientV2) BootstrapConnect(ctx context.Context) (fingerprint string, certPEM []byte, err error) {
	tr, ok := c.base.client.Transport.(*http.Transport)
	if !ok {
		return "", nil, fmt.Errorf("failed to assert client transport as *http.Transport")
	}

	tr.TLSClientConfig.InsecureSkipVerify = true

	opts := NewRequestOptions(&c.base.settings).WithURL("/auth/status").WithMethod(GET)
	_, resp, err := c.base.Do(ctx, opts)
	if err != nil && resp == nil {
		return "", nil, fmt.Errorf("initial connection failed: %w", err)
	}

	if resp.TLS == nil || len(resp.TLS.PeerCertificates) == 0 {
		return "", nil, fmt.Errorf("no TLS certificates presented by server")
	}

	rootCert := resp.TLS.PeerCertificates[len(resp.TLS.PeerCertificates)-1]

	sum := sha256.Sum256(rootCert.Raw)
	fingerprint = hex.EncodeToString(sum[:])

	certPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: rootCert.Raw,
	})

	return fingerprint, certPEM, nil
}

// ========MASTER-ENDPOINTS=========

func (c *ClientV2) GetAuthStatus(ctx context.Context) (*models.AuthStatusResponse, error) {
	opts := NewRequestOptions(&c.base.settings).
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

func (c *ClientV2) GetNodes(ctx context.Context) (*models.GetNodesResponse, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL("/cluster/nodes").
		WithMethod(GET)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.GetNodesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resp, nil
}

// ─────────────────────── User management ────────────────────────────────────

func (c *ClientV2) ListUsers(ctx context.Context) (*models.ListUsersResponse, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL("/auth/users").
		WithMethod(GET)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.ListUsersResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) GetUser(ctx context.Context, userID string) (*models.UserInfo, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/auth/users/%s", userID)).
		WithMethod(GET)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.UserInfo
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) DeleteUser(ctx context.Context, userID string) error {
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/auth/users/%s", userID)).
		WithMethod(DELETE)

	_, _, err := c.Do(ctx, opts)
	return err
}

func (c *ClientV2) GenerateUserToken(ctx context.Context, userID string) (string, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/auth/users/%s/token", userID)).
		WithMethod(POST)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return "", err
	}

	var resp map[string]string
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	return resp["token"], nil
}

func (c *ClientV2) AssignRole(ctx context.Context, userID string, req models.AssignRoleRequest) (*models.UserInfo, error) {
	data, _ := json.Marshal(req)
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/auth/users/%s/roles", userID)).
		WithMethod(POST).
		WithData(string(data))

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.UserInfo
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) RevokeToken(ctx context.Context, req models.RevokeTokenRequest) error {
	data, _ := json.Marshal(req)
	opts := NewRequestOptions(&c.base.settings).
		WithURL("/auth/tokens/revoke").
		WithMethod(POST).
		WithData(string(data))

	_, _, err := c.Do(ctx, opts)
	return err
}

func (c *ClientV2) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.UserInfo, error) {
	data, _ := json.Marshal(req)
	opts := NewRequestOptions(&c.base.settings).
		WithURL("/auth/users").
		WithMethod(POST).
		WithData(string(data))

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.UserInfo
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}

// ─────────────────────── Role management ────────────────────────────────────

func (c *ClientV2) ListRoles(ctx context.Context) (*models.ListRolesResponse, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL("/auth/roles").
		WithMethod(GET)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.ListRolesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) GetRole(ctx context.Context, roleName string) (*models.RoleInfo, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/auth/roles/%s", roleName)).
		WithMethod(GET)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.RoleInfo
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) CreateRole(ctx context.Context, req models.CreateRoleRequest) (*models.RoleInfo, error) {
	data, _ := json.Marshal(req)
	opts := NewRequestOptions(&c.base.settings).
		WithURL("/auth/roles").
		WithMethod(POST).
		WithData(string(data))

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.RoleInfo
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) UpdateRole(ctx context.Context, roleName string, req models.UpdateRoleRequest) (*models.RoleInfo, error) {
	data, _ := json.Marshal(req)
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/auth/roles/%s", roleName)).
		WithMethod(PUT).
		WithData(string(data))

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.RoleInfo
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) DeleteRole(ctx context.Context, roleName string) error {
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/auth/roles/%s", roleName)).
		WithMethod(DELETE)

	_, _, err := c.Do(ctx, opts)
	return err
}

// ========FILESTORE-ENDPOINTS=========

func (c *ClientV2) InitUpload(ctx context.Context, req *models.InitUploadRequest) (*models.InitUploadResponse, error) {
	if req.BucketID == "" {
		req.BucketID = models.GlobalBucketID
	}
	data, _ := json.Marshal(req)
	opts := NewRequestOptions(&c.base.settings).
		WithURL("/files").
		WithMethod(POST).
		WithData(string(data))

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.InitUploadResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) GetUploadStatus(ctx context.Context, fileUUID string) (*models.GetUploadStatusResponse, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/files/%s/status", fileUUID)).
		WithMethod(GET)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.GetUploadStatusResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode status: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) StreamUpload(ctx context.Context, fileUUID string, content io.Reader, offset int64) (*models.FinalizeUploadResponse, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/files/%s/content", fileUUID)).
		WithMethod(PUT).
		WithStream()
	if offset > 0 {
		opts.header.Set("Content-Range", fmt.Sprintf("bytes %d-", offset))
	}

	fullURL := buildURL(c.base.settings.BaseURL, opts)
	req, err := http.NewRequestWithContext(ctx, "PUT", fullURL, content)
	if err != nil {
		return nil, err
	}
	req.Header = opts.header

	resp, err := c.base.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, c.parseAPIError(body, resp.StatusCode)
	}

	var finalResp models.FinalizeUploadResponse
	if err := json.Unmarshal(body, &finalResp); err != nil {
		return nil, fmt.Errorf("failed to decode final response: %w", err)
	}
	return &finalResp, nil
}

func (c *ClientV2) UploadFileWrapper(ctx context.Context, filePath string, req *models.InitUploadRequest) (*models.FinalizeUploadResponse, error) {
	file, err := os.Open(filePath) // #nosec G304
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	req.Filename = filepath.Base(filePath)
	req.SizeBytes = stat.Size()
	if req.ContentType == "" {
		req.ContentType = "application/octet-stream"
	}

	initRes, err := c.InitUpload(ctx, req)
	if err != nil {
		return nil, err
	}

	status, err := c.GetUploadStatus(ctx, initRes.FileUUID)
	if err != nil {
		return nil, err
	}

	if status.Offset >= stat.Size() {
		return &models.FinalizeUploadResponse{
			FileUUID: initRes.FileUUID,
			Status:   "available",
		}, nil
	}

	if status.Offset > 0 {
		if _, err := file.Seek(status.Offset, io.SeekStart); err != nil {
			return nil, fmt.Errorf("failed to seek local file: %w", err)
		}
	}

	return c.StreamUpload(ctx, initRes.FileUUID, file, status.Offset)
}

func (c *ClientV2) DownloadFile(ctx context.Context, fileUUID, outputPath string) error {
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/files/%s", fileUUID)).
		WithMethod(GET).
		WithStream()

	_, resp, err := c.base.Do(ctx, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return c.parseAPIError(body, resp.StatusCode)
	}

	outFile, err := os.Create(outputPath) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (c *ClientV2) CreateBucket(ctx context.Context, req models.CreateBucketRequest) (*models.CreateBucketResponse, error) {
	data, _ := json.Marshal(req)
	opts := NewRequestOptions(&c.base.settings).
		WithURL("/buckets").
		WithMethod(POST).
		WithData(string(data))

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.CreateBucketResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) ListBucketFiles(ctx context.Context, bucketID string) (*models.ListBucketFilesResponse, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/buckets/%s/files", bucketID)).
		WithMethod(GET)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.ListBucketFilesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) DeleteFile(ctx context.Context, fileUUID string) (*models.DeleteFileResponse, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/files/%s", fileUUID)).
		WithMethod(DELETE)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.DeleteFileResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) UploadFileTobucketWrapper(ctx context.Context, bucketID, filePath string, req *models.InitUploadRequest) (*models.FinalizeUploadResponse, error) {
	file, err := os.Open(filePath) // #nosec G304
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	req.Filename = filepath.Base(filePath)
	req.SizeBytes = stat.Size()
	if req.ContentType == "" {
		req.ContentType = "application/octet-stream"
	}

	initRes, err := c.initUploadToBucket(ctx, bucketID, req)
	if err != nil {
		return nil, err
	}

	status, err := c.GetUploadStatus(ctx, initRes.FileUUID)
	if err != nil {
		return nil, err
	}

	if status.Offset >= stat.Size() {
		return &models.FinalizeUploadResponse{
			FileUUID: initRes.FileUUID,
			Status:   "available",
		}, nil
	}

	if status.Offset > 0 {
		if _, err := file.Seek(status.Offset, io.SeekStart); err != nil {
			return nil, fmt.Errorf("failed to seek local file: %w", err)
		}
	}

	return c.StreamUpload(ctx, initRes.FileUUID, file, status.Offset)
}

func (c *ClientV2) initUploadToBucket(ctx context.Context, bucketID string, req *models.InitUploadRequest) (*models.InitUploadResponse, error) {
	data, _ := json.Marshal(req)
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/buckets/%s/files", bucketID)).
		WithMethod(POST).
		WithData(string(data))

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.InitUploadResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) GeneratePublicToken(ctx context.Context, req *models.PublicTokenRequest) (*models.PublicTokenResponse, error) {
	data, _ := json.Marshal(req)
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/files/%s/public-token", req.FileUUID)).
		WithMethod(POST).
		WithData(string(data))

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.PublicTokenResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) DeleteBucket(ctx context.Context, bucketID string) (*models.DeleteBucketResponse, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/buckets/%s", bucketID)).
		WithMethod(DELETE)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.DeleteBucketResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode buckets: %w", err)
	}

	return &resp, nil
}

func (c *ClientV2) ListBuckets(ctx context.Context) (*models.ListBucketResponse, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL("/buckets").
		WithMethod(GET)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.ListBucketResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode buckets: %w", err)
	}

	return &resp, nil
}

// ========JOB-ENDPOINTS=========

func (c *ClientV2) ListJobs(ctx context.Context) (*models.ListJobsResponse, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL("/jobs").
		WithMethod(GET)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.ListJobsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode jobs: %w", err)
	}

	return &resp, nil
}

func (c *ClientV2) RunJob(ctx context.Context, jobName string, req *models.RunJobRequest) (*models.RunJobResponse, error) {
	data, _ := json.Marshal(req)
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/jobs/%s/run", jobName)).
		WithMethod(POST).
		WithData(string(data))

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.RunJobResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode run job response: %w", err)
	}

	return &resp, nil
}

func (c *ClientV2) GetJobRun(ctx context.Context, runID string) (*models.JobRunStatus, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/jobs/runs/%s", runID)).
		WithMethod(GET)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.JobRunStatus
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode job run: %w", err)
	}

	return &resp, nil
}

func (c *ClientV2) ListJobRuns(ctx context.Context) (*models.ListJobRunsResponse, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL("/jobs/runs").
		WithMethod(GET)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.ListJobRunsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode job runs: %w", err)
	}

	return &resp, nil
}

// ======== BUCKET PERMISSION ENDPOINTS ========

func (c *ClientV2) GrantBucketPermission(ctx context.Context, bucketID string, req models.GrantPermissionRequest) (*models.PermissionEntry, error) {
	data, _ := json.Marshal(req)
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/buckets/%s/permissions", bucketID)).
		WithMethod(POST).
		WithData(string(data))

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.PermissionEntry
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode permission response: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) RevokeBucketPermission(ctx context.Context, bucketID, permID string) (*models.RevokePermissionResponse, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/buckets/%s/permissions/%s", bucketID, permID)).
		WithMethod(DELETE)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.RevokePermissionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode revoke response: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) ListBucketPermissions(ctx context.Context, bucketID string) (*models.ListPermissionsResponse, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/buckets/%s/permissions", bucketID)).
		WithMethod(GET)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.ListPermissionsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode permissions list: %w", err)
	}
	return &resp, nil
}

// ======== FILE PERMISSION ENDPOINTS ========

func (c *ClientV2) GrantFilePermission(ctx context.Context, fileUUID string, req models.GrantPermissionRequest) (*models.PermissionEntry, error) {
	data, _ := json.Marshal(req)
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/files/%s/permissions", fileUUID)).
		WithMethod(POST).
		WithData(string(data))

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.PermissionEntry
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode permission response: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) RevokeFilePermission(ctx context.Context, fileUUID, permID string) (*models.RevokePermissionResponse, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/files/%s/permissions/%s", fileUUID, permID)).
		WithMethod(DELETE)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.RevokePermissionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode revoke response: %w", err)
	}
	return &resp, nil
}

func (c *ClientV2) ListFilePermissions(ctx context.Context, fileUUID string) (*models.ListPermissionsResponse, error) {
	opts := NewRequestOptions(&c.base.settings).
		WithURL(fmt.Sprintf("/files/%s/permissions", fileUUID)).
		WithMethod(GET)

	body, _, err := c.Do(ctx, opts)
	if err != nil {
		return nil, err
	}

	var resp models.ListPermissionsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode permissions list: %w", err)
	}
	return &resp, nil
}
