# GNS3UTIL Cluster Master

The master is the control plane for cluster auth, cluster state, node registration, and job orchestration.

## What It Does

- serves the HTTPS control-plane API
- serves dRPC for filestore-to-master coordination
- owns the etcd-backed RBAC and cluster state
- mints JWTs for users already present in etcd
- registers and dispatches cluster jobs
- bootstraps filestore nodes into the cluster

## Current API Surface

Main route groups in [`cmd/master/main.go`](/home/veya/coding/gns3util/cmd/master/main.go):

- `/api/v1/auth`
  - `POST /token`
  - `GET /status`
  - admin-only user, role, grant, revoke, and token-revoke endpoints
- `/api/v1/cluster`
  - cluster join
  - filestore join
  - node listing
- `/api/v1/jobs`
  - list jobs
  - run job
  - list job runs
  - get job run

## RBAC Status

The master/state RBAC path is implemented and is ahead of the old board notes.

Done now:

- role-based permission checks are enforced through `RequireScope`
- built-in `admin` resolves to effective `admin:*`
- `GET /api/v1/auth/status` works for any authenticated bearer, not just admins
- auth status and user responses return typed effective permissions
- explicit user deny scopes are supported in etcd state
- deny takes precedence over role allows, including built-in admin
- precedence behavior is covered in [`pkg/state/state_test.go`](/home/veya/coding/gns3util/pkg/state/state_test.go)
- CLI support exists for inspecting effective permissions and creating users with deny scopes

Important boundary:

- explicit deny is implemented in the master/state RBAC layer
- bucket/file ACLs in the filestore are still a separate allow-oriented layer; they are not yet a full deny-aware policy engine

## Config

Environment variables are defined by `MasterConfig` in [`cmd/master/main.go`](/home/veya/coding/gns3util/cmd/master/main.go):

- `MASTER_TLS_DIR`
- `MASTER_API_PORT`
- `MASTER_STORE_DRPC_PORT`
- `MASTER_API_LISTEN_ADDR`
- `MASTER_ETCD_API_PORT`
- `MASTER_ETCD_API_LISTEN_ADDR`
- `CLUSTER_PRIV_KEY`
- `MASTER_TLS_SUBJ`
- `MASTER_DATA_DIR`
- `MASTER_ENABLE_MDNS`
- `OTEL_ENDPOINT`
- `APP_NAME`
- `MASTER_ADVERTISE_ADDR`
- `METRICS_ENABLED`

## Notes Against The Object-Storage Board

The codebase now covers these RBAC items even if the board still shows them as open:

- explicit deny rules for users
- RBAC tests for overwrite/precedence behavior

The remaining RBAC item that still applies conceptually is the filestore-side authorization redesign:

- bucket authorization / ACL hierarchy cleanup is still a filestore concern, not a missing master-state RBAC feature

## Docs

- Swagger UI: `/swagger`
- ReDoc: `/redoc`
