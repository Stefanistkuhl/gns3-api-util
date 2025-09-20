package exercise

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/cluster/db"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
)

func NewExerciseLsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List exercises distribution across nodes",
		RunE:  runExerciseLs,
	}
	cmd.Flags().Bool("db-only", false, "Use only DB for listing")
	cmd.Flags().Bool("api-only", false, "Use only API for listing (not implemented)")
	cmd.Flags().StringP("cluster", "c", "", "Cluster name")
	cmd.Flags().String("class", "", "Filter by class name")
	return cmd
}

func runExerciseLs(cmd *cobra.Command, args []string) error {
	cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get global options: %w", err)
	}
	clusterName, _ := cmd.Flags().GetString("cluster")
	className, _ := cmd.Flags().GetString("class")

	clusterID, err := resolveExerciseClusterID(cfg, clusterName)
	if err != nil {
		return err
	}

	conn, err := db.InitIfNeeded()
	if err != nil {
		return fmt.Errorf("failed to init db: %w", err)
	}
	defer conn.Close()

	nodes, err := db.GetNodeExercisesForClass(conn, clusterID, className)
	if err != nil {
		return fmt.Errorf("failed to get exercise distribution: %w", err)
	}

	type row struct{ Node, Exercises, Projects, Groups string }
	rows := make([]row, 0, len(nodes))
	for _, n := range nodes {
		exNames := make(map[string]bool)
		projects := make(map[string]bool)
		groups := make(map[string]bool)
		for _, it := range n.Exercises {
			exNames[it.Name] = true
			projects[it.ProjectUUID] = true
			groups[it.GroupName] = true
		}
		rows = append(rows, row{
			Node:      n.NodeURL,
			Exercises: fmt.Sprintf("%d", len(exNames)),
			Projects:  fmt.Sprintf("%d", len(projects)),
			Groups:    fmt.Sprintf("%d", len(groups)),
		})
	}

	utils.PrintTable(rows, []utils.Column[row]{
		{Header: "Node", Value: func(r row) string { return r.Node }},
		{Header: "Exercises", Value: func(r row) string { return r.Exercises }},
		{Header: "Projects", Value: func(r row) string { return r.Projects }},
		{Header: "Groups", Value: func(r row) string { return r.Groups }},
	})

	for _, n := range nodes {
		if len(n.Exercises) == 0 {
			continue
		}
		fmt.Printf("\n%s %s\n", messageUtils.Bold("Node:"), messageUtils.Highlight(n.NodeURL))
		fmt.Println(strings.Repeat("-", 69))
		sort.Slice(n.Exercises, func(i, j int) bool { return n.Exercises[i].Name < n.Exercises[j].Name })
		for _, it := range n.Exercises {
			name := it.Name
			grp := it.GroupName
			uuid := it.ProjectUUID
			state := it.State
			if len(name) > 80 {
				name = name[:77] + "..."
			}
			if len(grp) > 80 {
				grp = grp[:77] + "..."
			}
			if len(uuid) > 80 {
				uuid = uuid[:77] + "..."
			}
			fmt.Printf("  - %s  %s %s  %s %s  %s %s\n",
				messageUtils.Bold(name),
				messageUtils.Highlight("group:"), grp,
				messageUtils.Highlight("project:"), uuid,
				messageUtils.Highlight("state:"), state,
			)
		}
		fmt.Println()
	}

	return nil
}

func resolveExerciseClusterID(cfg config.GlobalOptions, clusterName string) (int, error) {
	conn, err := db.InitIfNeeded()
	if err != nil {
		return 0, fmt.Errorf("failed to init db: %w", err)
	}
	defer conn.Close()

	if clusterName != "" {
		clusters, err := db.GetClusters(conn)
		if err != nil {
			return 0, fmt.Errorf("failed to get clusters: %w", err)
		}
		for _, c := range clusters {
			if strings.EqualFold(strings.TrimSpace(c.Name), strings.TrimSpace(clusterName)) {
				return c.Id, nil
			}
		}
		return 0, fmt.Errorf("cluster not found: %s", clusterName)
	}

	if cfg.Server == "" {
		return 0, fmt.Errorf("no server configured; use -s or provide -c cluster name")
	}
	u := utils.ValidateUrlWithReturn(cfg.Server)
	if u == nil {
		return 0, fmt.Errorf("invalid server url: %s", cfg.Server)
	}
	derived := fmt.Sprintf("%s%s", u.Hostname(), "_single_node_cluster")
	clusters, err := db.GetClusters(conn)
	if err != nil {
		return 0, fmt.Errorf("failed to get clusters: %w", err)
	}
	for _, c := range clusters {
		if c.Name == derived {
			return c.Id, nil
		}
	}
	return 0, fmt.Errorf("cluster not found: %s", derived)
}
