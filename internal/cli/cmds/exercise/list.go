package exercise

import (
	"fmt"
	"sort"
	"strings"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	clusterutils "github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/clusterUtils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/messageUtils"
	"github.com/spf13/cobra"
)

func NewExerciseLsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List exercises distribution across nodes",
		RunE:  runExerciseLs,
	}
	cmd.Flags().Bool("db-only", false, "Use only DB for listing")
	cmd.Flags().Bool("api-only", false, "Use only API for listing (not implemented)")
	cmd.Flags().StringP("cluster", "", "", "Cluster name")
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

	clusterID, err := clusterutils.ResolveClusterID(cfg, clusterName, cmd.Context())
	if err != nil {
		return err
	}

	store, err := db.Init()
	if err != nil {
		return fmt.Errorf("failed to init db: %w", err)
	}

	sqlcRows, err := store.GetNodeExercisesForCluster(cmd.Context(), int64(clusterID))
	if err != nil {
		return fmt.Errorf("failed to get exercise distribution: %w", err)
	}

	nodes := clusterutils.TransformNodeExercises(sqlcRows, className)

	if len(nodes) == 0 {
		fmt.Println(messageUtils.InfoMsg("No exercises found in DB for this cluster"))
		return nil
	}

	type row struct{ Node, Exercises, Projects, Groups string }
	tableRows := make([]row, 0, len(nodes))
	for _, n := range nodes {
		exNames := make(map[string]bool)
		projects := make(map[string]bool)
		groups := make(map[string]bool)
		for _, it := range n.Exercises {
			exNames[it.Name] = true
			projects[it.ProjectUUID] = true
			groups[it.GroupName] = true
		}
		tableRows = append(tableRows, row{
			Node:      n.NodeURL,
			Exercises: fmt.Sprintf("%d", len(exNames)),
			Projects:  fmt.Sprintf("%d", len(projects)),
			Groups:    fmt.Sprintf("%d", len(groups)),
		})
	}

	utils.PrintTable(tableRows, []utils.Column[row]{
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
