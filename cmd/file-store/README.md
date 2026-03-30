# GNS3UTIL File Store

The filestore is the object-storage node for the cluster. It serves uploads/downloads, bucket and file ACL APIs, public file tokens, and node-local maintenance jobs.

## What It Does

- stores blobs on local disk
- stores metadata and ACLs in the local SQLite/Turso-backed DB
- exposes bucket, file, public-download, and job endpoints over HTTPS/HTTP3
- asks the master RBAC service for global scope checks
- enforces bucket/file ACL checks locally
- runs cleanup and vacuum work in the background

## Current API Surface

Main route groups in [`cmd/file-store/main.go`](/home/veya/coding/gns3util/cmd/file-store/main.go):

- `/api/v1/buckets`
  - list buckets
  - create bucket
  - list bucket files
  - init upload into a bucket
  - delete bucket
  - bucket permission CRUD
- `/api/v1/files`
  - init upload without a bucket path
  - upload status
  - download
  - stream content upload
  - delete file
  - generate public token
  - file permission CRUD
- `/api/v1/public`
  - public download by token
- `/api/v1/jobs`
  - run node-local jobs such as `db-vacuum`

## Access Model

The filestore uses two layers:

1. Master RBAC
- global cluster scopes like `read:files`, `write:files`, `execute:jobs`, `read:metrics`
- explicit deny support now lives in the master/state layer and is enforced before allow

2. Filestore-local ACLs
- bucket permissions: `read`, `write`, `admin`
- file permissions: `read`, `write`, `admin`
- file access inherits from the parent bucket where applicable

## RBAC / ACL Status

What is done:

- bucket and file routes are guarded by both global RBAC and local ACL middleware
- file routes honor bucket inheritance instead of owner-only behavior
- delegated bucket/file admins work for delete, upload status, content upload, and public-token flows
- permission-management endpoints respect global admin and inherited bucket admin
- file ACL behavior has targeted tests in [`rbac_test.go`](/home/veya/coding/gns3util/internal/file-store/handlers/rbac_test.go)

What is still not a finished policy model:

- bucket/file ACLs are still allow-based
- there is not yet a full deny-aware bucket/file policy hierarchy
- bucket `required_scopes` is still a compatibility mechanism and not the final authorization model

That matches the remaining filestore-side RBAC design work on the object-storage board.

## Config

Environment variables are defined by `FilestoreConfig` in [`cmd/file-store/main.go`](/home/veya/coding/gns3util/cmd/file-store/main.go):

- `FILE_STORE_NODE_NAME`
- `FILE_STORE_API_PORT`
- `FILE_STORE_DRPC_PORT`
- `FILE_STORE_API_LISTEN_ADDR`
- `FILE_STORE_ADVERTISE_ADDR`
- `MASTER_API_URL`
- `MASTER_DRPC_ADDR`
- `CLUSTER_PUB_KEY`
- `FILE_STORE_TLS_DIR`
- `OTEL_ENDPOINT`
- `APP_NAME`
- `JOIN_TOKEN`
- `FILE_STORE_DB_PATH`
- `FILE_STORE_DATA_PATH`
- `FILE_STORE_VACUUM_INTERVAL`
- `FILE_STORE_TOMBSTONE_EXPIRY`
- `FILE_STORE_PENDING_EXPIRY`
- `FILE_STORE_STALLED_UPLOAD_EXPIRY`
- `METRICS_ENABLED`

## Notes Against The Object-Storage Board

Board items that are now effectively done in code:

- RBAC hardening for bucket/file operations
- test coverage for current filestore ACL behavior

Board items that still remain real for this component:

- bucket authorization and permission-model redesign
- backup/project-file endpoints
- VM/project-file metadata parsing on ingest
- startup crash-recovery indexing for orphaned tmp files
- inter-node object-store RPC methods

## Docs

- Swagger UI: `/swagger`
- ReDoc: `/redoc`
