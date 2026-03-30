package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/pkg/api"
	"github.com/0xveya/gns3util/pkg/models"
)

func newFilestoreClient(
	serverURL,
	token string,
	cfg *config.GlobalOptions,
) *api.ClientV2 {
	settings := api.NewSettings(
		api.WithBaseURLV2(serverURL+"/api/v1"),
		api.WithToken(token),
		api.WithVerify(!cfg.Insecure),
		api.WithCA([]byte(cfg.ClusterEntry.CaCert)),
		api.WithHTTP3(true),
	)
	return api.NewClientV2(&settings)
}

func RunUpload(
	serverURL,
	token,
	filePath string,
	reqOpts *models.InitUploadRequest,
	cfg *config.GlobalOptions,
) (*models.FinalizeUploadResponse, error) {
	if reqOpts.BucketID == "" {
		reqOpts.BucketID = models.GlobalBucketID
	}
	client := newFilestoreClient(serverURL, token, cfg)
	ctx := context.Background()

	resp, err := client.UploadFileWrapper(ctx, filePath, reqOpts)
	if err != nil {
		return nil, fmt.Errorf("upload execution failed: %w", err)
	}

	return resp, nil
}

func RunDownload(
	serverURL,
	token,
	fileUUID,
	outputPath string,
	cfg *config.GlobalOptions,
) error {
	client := newFilestoreClient(serverURL, token, cfg)
	ctx := context.Background()

	err := client.DownloadFile(ctx, fileUUID, outputPath)
	if err != nil {
		return fmt.Errorf("download execution failed: %w", err)
	}

	return nil
}

func RunCreateBucket(
	serverURL,
	token,
	bucketName string,
	isPublic bool,
	requiredScopes string,
	cfg *config.GlobalOptions,
) (*models.CreateBucketResponse, error) {
	client := newFilestoreClient(serverURL, token, cfg)
	ctx := context.Background()

	req := models.CreateBucketRequest{
		Name:           bucketName,
		IsPublic:       isPublic,
		RequiredScopes: &requiredScopes,
	}

	resp, err := client.CreateBucket(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create bucket failed: %w", err)
	}

	return resp, nil
}

func RunListBucketFiles(
	serverURL,
	token,
	bucketID string,
	cfg *config.GlobalOptions,
) (*models.ListBucketFilesResponse, error) {
	client := newFilestoreClient(serverURL, token, cfg)
	ctx := context.Background()

	files, err := client.ListBucketFiles(ctx, bucketID)
	if err != nil {
		return nil, fmt.Errorf("list bucket files failed: %w", err)
	}

	return files, nil
}

func RunDeleteFile(
	serverURL,
	token,
	fileUUID string,
	cfg *config.GlobalOptions,
) (*models.DeleteFileResponse, error) {
	client := newFilestoreClient(serverURL, token, cfg)
	ctx := context.Background()

	resp, err := client.DeleteFile(ctx, fileUUID)
	if err != nil {
		return nil, fmt.Errorf("delete file failed: %w", err)
	}

	return resp, nil
}

func RunUploadToBucket(
	serverURL,
	token,
	bucketID,
	filePath string,
	reqOpts *models.InitUploadRequest,
	cfg *config.GlobalOptions,
) (*models.FinalizeUploadResponse, error) {
	client := newFilestoreClient(serverURL, token, cfg)
	ctx := context.Background()

	resp, err := client.UploadFileTobucketWrapper(
		ctx,
		bucketID,
		filePath,
		reqOpts,
	)
	if err != nil {
		return nil, fmt.Errorf("upload to bucket failed: %w", err)
	}

	return resp, nil
}

func RunGeneratePublicToken(
	serverURL,
	token,
	fileUUID,
	bucketID string,
	expiresAt *time.Time,
	cfg *config.GlobalOptions,
) (*models.PublicTokenResponse, error) {
	client := newFilestoreClient(serverURL, token, cfg)
	ctx := context.Background()
	req := &models.PublicTokenRequest{
		FileUUID:   fileUUID,
		BucketUUID: bucketID,
		ExpiresAt:  expiresAt,
	}

	resp, err := client.GeneratePublicToken(
		ctx,
		req,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"generate public token failed: %w",
			err,
		)
	}

	return resp, nil
}

func RunGetUploadStatus(
	serverURL,
	token,
	fileUUID string,
	cfg *config.GlobalOptions,
) (*models.GetUploadStatusResponse, error) {
	client := newFilestoreClient(serverURL, token, cfg)
	ctx := context.Background()

	status, err := client.GetUploadStatus(ctx, fileUUID)
	if err != nil {
		return nil, fmt.Errorf("get upload status failed: %w", err)
	}

	return status, nil
}

func RunDeleteBucket(
	serverURL,
	token,
	bucketID string,
	cfg *config.GlobalOptions,
) (*models.DeleteBucketResponse, error) {
	client := newFilestoreClient(serverURL, token, cfg)
	ctx := context.Background()

	resp, err := client.DeleteBucket(ctx, bucketID)
	if err != nil {
		return nil, fmt.Errorf("delete bucket failed: %w", err)
	}

	return resp, nil
}

func RunListBuckets(
	serverURL,
	token string,
	cfg *config.GlobalOptions,
) (*models.ListBucketResponse, error) {
	client := newFilestoreClient(serverURL, token, cfg)
	ctx := context.Background()

	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("list buckets failed: %w", err)
	}

	return buckets, nil
}

// ─────────────────────────── Bucket permission helpers ───────────────────────

func RunGrantBucketPermission(
	serverURL, token, bucketID string,
	req models.GrantPermissionRequest,
	cfg *config.GlobalOptions,
) (*models.PermissionEntry, error) {
	client := newFilestoreClient(serverURL, token, cfg)
	resp, err := client.GrantBucketPermission(context.Background(), bucketID, req)
	if err != nil {
		return nil, fmt.Errorf("grant bucket permission failed: %w", err)
	}
	return resp, nil
}

func RunRevokeBucketPermission(
	serverURL, token, bucketID, permID string,
	cfg *config.GlobalOptions,
) (*models.RevokePermissionResponse, error) {
	client := newFilestoreClient(serverURL, token, cfg)
	resp, err := client.RevokeBucketPermission(context.Background(), bucketID, permID)
	if err != nil {
		return nil, fmt.Errorf("revoke bucket permission failed: %w", err)
	}
	return resp, nil
}

func RunListBucketPermissions(
	serverURL, token, bucketID string,
	cfg *config.GlobalOptions,
) (*models.ListPermissionsResponse, error) {
	client := newFilestoreClient(serverURL, token, cfg)
	resp, err := client.ListBucketPermissions(context.Background(), bucketID)
	if err != nil {
		return nil, fmt.Errorf("list bucket permissions failed: %w", err)
	}
	return resp, nil
}

// ─────────────────────────── File permission helpers ─────────────────────────

func RunGrantFilePermission(
	serverURL, token, fileUUID string,
	req models.GrantPermissionRequest,
	cfg *config.GlobalOptions,
) (*models.PermissionEntry, error) {
	client := newFilestoreClient(serverURL, token, cfg)
	resp, err := client.GrantFilePermission(context.Background(), fileUUID, req)
	if err != nil {
		return nil, fmt.Errorf("grant file permission failed: %w", err)
	}
	return resp, nil
}

func RunRevokeFilePermission(
	serverURL, token, fileUUID, permID string,
	cfg *config.GlobalOptions,
) (*models.RevokePermissionResponse, error) {
	client := newFilestoreClient(serverURL, token, cfg)
	resp, err := client.RevokeFilePermission(context.Background(), fileUUID, permID)
	if err != nil {
		return nil, fmt.Errorf("revoke file permission failed: %w", err)
	}
	return resp, nil
}

func RunListFilePermissions(
	serverURL, token, fileUUID string,
	cfg *config.GlobalOptions,
) (*models.ListPermissionsResponse, error) {
	client := newFilestoreClient(serverURL, token, cfg)
	resp, err := client.ListFilePermissions(context.Background(), fileUUID)
	if err != nil {
		return nil, fmt.Errorf("list file permissions failed: %w", err)
	}
	return resp, nil
}
