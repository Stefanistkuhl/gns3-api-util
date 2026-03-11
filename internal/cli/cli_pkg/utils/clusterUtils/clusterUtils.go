package clusterutils

import (
	"context"
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db/sqlc"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
)

func ResolveClusterID(cfg *config.GlobalOptions, clusterName string, ctx context.Context) (int, error) {
	store, err := db.Init()
	if err != nil {
		return 0, fmt.Errorf("failed to init db: %w", err)
	}

	if clusterName != "" {
		clusters, getClustersErr := store.GetClusters(context.Background())
		if getClustersErr != nil {
			return 0, fmt.Errorf("failed to get clusters: %w", getClustersErr)
		}
		for _, c := range clusters {
			if c.Name == clusterName {
				return int(c.ClusterID), nil
			}
		}
		return 0, fmt.Errorf("cluster not found: %s", clusterName)
	}

	if cfg.Server == "" {
		return 0, fmt.Errorf("no server configured; use -s or provide -c cluster name")
	}
	urlObj := utils.ValidateURLWithReturn(cfg.Server)
	if urlObj == nil {
		return 0, fmt.Errorf("invalid server url: %s", cfg.Server)
	}
	derived := fmt.Sprintf("%s%s", urlObj.Hostname(), "_single_node_cluster")
	clusters, err := store.GetClusters(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to get clusters: %w", err)
	}
	for _, c := range clusters {
		if c.Name == derived {
			return int(c.ClusterID), nil
		}
	}
	return 0, fmt.Errorf("cluster not found: %s", derived)
}

func TruncateList(items []string, maxLen int) []string {
	out := make([]string, len(items))
	for i, s := range items {
		if maxLen > 3 && len(s) > maxLen {
			out[i] = s[:maxLen-3] + "..."
		} else {
			out[i] = s
		}
	}
	return out
}

func TransformNodeExercises(rows []sqlc.GetNodeExercisesForClusterRow, className string) []db.NodeExercisesForClass {
	filtered := make([]sqlc.GetNodeExercisesForClusterRow, 0)
	for _, row := range rows {
		if className == "" || row.Name == className {
			filtered = append(filtered, row)
		}
	}

	results := make([]db.NodeExercisesForClass, 0)
	var current *db.NodeExercisesForClass

	for _, row := range filtered {
		nodeURL, ok := row.NodeUrl.(string)
		if !ok {
			continue // skip invalid entries
		}
		if current == nil || current.NodeURL != nodeURL {
			if current != nil {
				results = append(results, *current)
			}
			current = &db.NodeExercisesForClass{
				NodeURL:   nodeURL,
				Exercises: make([]db.ExerciseItem, 0),
			}
		}

		state := ""
		if row.State.Valid {
			state = row.State.String
		}

		current.Exercises = append(current.Exercises, db.ExerciseItem{
			Name:        row.ExerciseName,
			ProjectUUID: row.ProjectUuid,
			GroupName:   row.GroupName,
			State:       state,
		})
	}
	if current != nil {
		results = append(results, *current)
	}

	return results
}
