package handlers

import "errors"

var (
	ErrFileNotFound    = errors.New("file not found")
	ErrFileUnavailable = errors.New("file not available")
)

var (
	ErrInvalidRange               = errors.New("invalid range")
	ErrMultipleRangesNotSupported = errors.New("multiple ranges not supported")
	ErrOpenFile                   = errors.New("open file failed")
	ErrSeekFile                   = errors.New("seek file failed")
	ErrCopyFile                   = errors.New("copy file failed")
)

var (
	ErrUploadOffsetMismatch    = errors.New("upload offset mismatch")
	ErrInvalidContentRange     = errors.New("invalid content-range")
	ErrMarkFileUploading       = errors.New("mark file uploading failed")
	ErrHashExistingPartialFile = errors.New("hash existing partial file failed")
	ErrOpenUploadTempFile      = errors.New("open upload temp file failed")
	ErrSeekUploadTempFile      = errors.New("seek upload temp file failed")
	ErrStreamUploadChunk       = errors.New("stream upload chunk failed")
	ErrSyncUploadTempFile      = errors.New("sync upload temp file failed")
	ErrCreateShardDirs         = errors.New("create shard dirs failed")
	ErrMoveUploadToBlob        = errors.New("move upload to blob failed")
	ErrFinalizeUploadDB        = errors.New("finalize upload db failed")
	ErrRemoveUploadTempFile    = errors.New("remove upload temp file failed")
	ErrorCloseUploadTempFile   = errors.New("close upload temp file failed")
	ErrorUpsertBlob            = errors.New("upsert blob failed")
	ErrorNoBucketID            = errors.New("bucket_id is required")
	ErrorNoPubToken            = errors.New("token is required, token in URL path")
	ErrorTokenNotFound         = errors.New("token not found")
	ErrorFailedToQueryDB       = errors.New("failed to query db")
	ErrorBucketIDMissmatch     = errors.New("bucket_id mismatch")
	ErrorExpiredToken          = errors.New("token expired")
	ErrorFileNotAvailable      = errors.New("file not available")
)
