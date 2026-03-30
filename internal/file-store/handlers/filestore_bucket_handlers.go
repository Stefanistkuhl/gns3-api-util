package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/utils/dbutils"
	"github.com/0xveya/gns3util/pkg/web/helpers"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// CreateBucket creates a new storage bucket
//
//	@Summary		Create bucket
//	@Description	Creates a new bucket for organizing files
//	@Tags			buckets
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.CreateBucketRequest	true	"Bucket creation request"
//	@Success		200		{object}	models.CreateBucketResponse
//	@Failure		400		{object}	helpers.APIErrorResponse
//	@Failure		401		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/buckets [post]
//	@Security		BearerAuth
func (f *FilestoreHandlers) CreateBucket(w http.ResponseWriter, r *http.Request) {
	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}
	userID := claims.UserID

	var req models.CreateBucketRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if req.Name == "" {
		helpers.WriteAPIError(w, "name is required", helpers.ErrCodeInvalidInput, "missing required field: name", http.StatusBadRequest)
		return
	}

	bucketUUID, uuidErr := uuid.NewV7()
	if uuidErr != nil {
		f.Logger.Error("Failed to generate UUID", "err", uuidErr, "user_id", userID)
		helpers.WriteAPIError(w, "failed to generate a uuid", helpers.ErrCodeFailedToGenerateUUID, "failed to generate a uuid for the bucket", http.StatusInternalServerError)
		return
	}

	params := sqlc_file_store.CreateBucketParams{
		BucketID:       bucketUUID.String(),
		Name:           req.Name,
		OwnerID:        userID,
		BucketType:     "standard",
		IsPublic:       dbutils.NullBool(&req.IsPublic),
		RequiredScopes: dbutils.NullString(req.RequiredScopes),
	}

	bucket, insertErr := f.Store.CreateBucket(r.Context(), params)
	if insertErr != nil {
		f.Logger.Error("Failed to create bucket", "err", insertErr, "user_id", userID, "name", req.Name)
		helpers.WriteAPIError(w, "failed to insert bucket into db", helpers.ErrCodeDBErr, insertErr.Error(), http.StatusInternalServerError)
		return
	}

	f.Logger.Info("Bucket created",
		"bucket_id", bucket.BucketID,
		"name", bucket.Name,
		"user_id", userID,
	)

	writeErr := helpers.WriteJSON(w, models.CreateBucketResponse{
		BucketID:       bucket.BucketID,
		Name:           bucket.Name,
		IsPublic:       bucket.IsPublic.Bool,
		RequiredScopes: bucket.RequiredScopes.String,
		CreatedAt:      dbutils.ParseDBTime(bucket.CreatedAt),
	})
	if writeErr != nil {
		f.Logger.Error("Failed to write response", "err", writeErr, "bucket_id", bucket.BucketID)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
}

// ListBucketFiles lists all files in a bucket
//
//	@Summary		List bucket files
//	@Description	Returns all files in a specific bucket
//	@Tags			buckets
//	@Produce		json
//	@Param			bucket_id	path		string	true	"Bucket ID"
//	@Success		200			{object}	models.ListBucketFilesResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/buckets/{bucket_id}/files [get]
//	@Security		BearerAuth
func (f *FilestoreHandlers) ListBucketFiles(w http.ResponseWriter, r *http.Request) {
	bucketID := chi.URLParam(r, "bucket_id")
	if bucketID == "" {
		helpers.WriteAPIError(w, "bucket_id is required", helpers.ErrCodeInvalidInput, "missing bucket_id in URL path", http.StatusBadRequest)
		return
	}

	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	files, queryErr := f.Store.ListFilesByBucketWithBlob(r.Context(), bucketID)
	if queryErr != nil {
		f.Logger.Error("Failed to list bucket files", "err", queryErr, "bucket_id", bucketID, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to query bucket files", helpers.ErrCodeDBErr, queryErr.Error(), http.StatusInternalServerError)
		return
	}

	fileResponses := make([]models.FileInfo, 0, len(files))
	for i := range files {
		file := files[i]
		fileResponses = append(fileResponses, models.FileInfo{
			FileUUID:    file.FileUuid,
			Filename:    file.Filename,
			SizeBytes:   file.SizeBytes,
			ContentType: file.ContentType,
			BlobSHA256:  file.BlobSha256.String,
			Status:      models.FileStatus(file.Status),
			CreatedAt:   dbutils.ParseDBTime(file.CreatedAt),
		})
	}

	writeErr := helpers.WriteJSON(w, models.ListBucketFilesResponse{BucketUUID: bucketID, Files: fileResponses, Count: len(fileResponses)})
	if writeErr != nil {
		f.Logger.Error("Failed to write response", "err", writeErr, "bucket_id", bucketID)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
}

// DeleteBucket deletes a bucket and all its files
//
//	@Summary		Delete bucket
//	@Description	Deletes a bucket and all files contained within it
//	@Tags			buckets
//	@Produce		json
//	@Param			bucket_id	path		string	true	"Bucket ID"
//	@Success		200			{object}	models.DeleteBucketResponse
//	@Failure		400			{object}	helpers.APIErrorResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/buckets/{bucket_id} [delete]
//	@Security		BearerAuth
func (f *FilestoreHandlers) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	bucketID := chi.URLParam(r, "bucket_id")
	if bucketID == "" {
		helpers.WriteAPIError(w, "bucket_id is required", helpers.ErrCodeInvalidInput, "", http.StatusBadRequest)
		return
	}

	if bucketID == models.GlobalBucketID {
		helpers.WriteAPIError(w, "Forbidden: Cannot delete system bucket", helpers.ErrCodeForbidden, "The global-default bucket is protected", http.StatusForbidden)
		return
	}

	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	bucket, ok := lookupBucketOrErr(r.Context(), w, f.Store, bucketID)
	if !ok {
		return
	}

	allowed, err := hasBucketPermission(r.Context(), f.Store, claims, &bucket, "admin")
	if err != nil {
		f.Logger.Error("bucket delete permission check failed", "bucket_id", bucketID, "err", err, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to verify bucket permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !allowed {
		helpers.WriteAPIError(w, "Forbidden", helpers.ErrCodeForbidden, "you do not have permission to delete this bucket", http.StatusForbidden)
		return
	}

	files, err := f.Store.ListFilesByBucketWithBlob(r.Context(), bucketID)
	if err != nil {
		f.Logger.Error("Failed to list bucket files for deletion", "bucket_id", bucketID, "err", err)
		helpers.WriteAPIError(w, "Internal error", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
		return
	}

	var blobsToRemove []string
	txErr := f.Store.WithTx(r.Context(), func(q *sqlc_file_store.Queries) error {
		for i := range files {
			if drefErr := q.DecrementBlobRefCount(r.Context(), files[i].BlobSha256.String); drefErr != nil {
				return fmt.Errorf("failed to decrement blob refcount for file %s: %w", files[i].FileUuid, drefErr)
			}

			updatedBlob, gErr := q.GetBlobBySHA256(r.Context(), files[i].BlobSha256.String)
			if gErr != nil {
				return fmt.Errorf("failed to get updated blob info for %s: %w", files[i].BlobSha256.String, gErr)
			}

			if updatedBlob.RefCount <= 0 {
				if bDelErr := q.DeleteBlob(r.Context(), files[i].BlobSha256.String); bDelErr != nil {
					return fmt.Errorf("failed to delete unreferenced blob %s: %w", files[i].BlobSha256.String, bDelErr)
				}
				blobsToRemove = append(blobsToRemove, updatedBlob.FilePath)
			}
		}

		if dErr := q.DeleteBucket(r.Context(), bucketID); dErr != nil {
			return fmt.Errorf("failed to delete bucket: %w", dErr)
		}
		return nil
	})

	if txErr != nil {
		f.Logger.Error("Failed to delete bucket in database", "bucket_id", bucketID, "err", txErr)
		helpers.WriteAPIError(w, "Internal error", helpers.ErrCodeDBErr, txErr.Error(), http.StatusInternalServerError)
		return
	}

	for _, path := range blobsToRemove {
		if rmErr := os.Remove(path); rmErr != nil {
			f.Logger.Warn("Failed to remove blob file from disk", "err", rmErr, "path", path)
		}
	}

	f.Logger.Info("Bucket deleted", "bucket_id", bucketID, "user_id", claims.UserID)

	writeErr := helpers.WriteJSON(w, models.DeleteBucketResponse{BucketUUID: bucketID, Status: "deleted"})

	if writeErr != nil {
		f.Logger.Error("Failed to write response", "err", writeErr, "bucket_uuid", bucketID)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
}

// ListBuckets lists all buckets owned by the user
//
//	@Summary		List buckets
//	@Description	Returns all buckets accessible to the authenticated user
//	@Tags			buckets
//	@Produce		json
//	@Success		200	{object}	models.ListBucketResponse
//	@Failure		401	{object}	helpers.APIErrorResponse
//	@Failure		500	{object}	helpers.APIErrorResponse
//	@Router			/api/v1/buckets [get]
//	@Security		BearerAuth
func (f *FilestoreHandlers) ListBuckets(w http.ResponseWriter, r *http.Request) {
	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	buckets, err := f.Store.ListBucketsAccessibleToUser(r.Context(), claims.UserID, roleNamesFromClaims(claims))
	if err != nil {
		f.Logger.Error("Failed to list buckets", "err", err, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "Internal error", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
		return
	}

	globalBucket, err := f.Store.GetBucketByID(r.Context(), models.GlobalBucketID)

	resp := make([]models.CreateBucketResponse, 0)

	if err == nil {
		resp = append(resp, models.CreateBucketResponse{
			BucketID:       globalBucket.BucketID,
			Name:           globalBucket.Name,
			IsPublic:       globalBucket.IsPublic.Bool,
			RequiredScopes: globalBucket.RequiredScopes.String,
			CreatedAt:      dbutils.ParseDBTime(globalBucket.CreatedAt),
		})
	}

	for i := range buckets {
		b := buckets[i]
		if b.BucketID == models.GlobalBucketID {
			continue
		}
		resp = append(resp, models.CreateBucketResponse{
			BucketID:       b.BucketID,
			Name:           b.Name,
			IsPublic:       b.IsPublic.Bool,
			RequiredScopes: b.RequiredScopes.String,
			CreatedAt:      dbutils.ParseDBTime(b.CreatedAt),
		})
	}

	writeErr := helpers.WriteJSON(w, models.ListBucketResponse{Buckets: resp, Count: len(resp)})
	if writeErr != nil {
		f.Logger.Error("Failed to write response", "err", writeErr)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
}
