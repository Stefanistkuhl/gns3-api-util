package exercise

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/cluster/db"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/fuzzy"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
)

func NewExerciseInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show detailed info for a single exercise across nodes",
		RunE:  runExerciseInfo,
	}
	cmd.Flags().StringP("cluster", "c", "", "Cluster name")
	return cmd
}

func runExerciseInfo(cmd *cobra.Command, args []string) error {
	cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get global options: %w", err)
	}
	clusterName, _ := cmd.Flags().GetString("cluster")

	clusterID, err := resolveExerciseClusterID(cfg, clusterName)
	if err != nil {
		return err
	}

	conn, err := db.InitIfNeeded()
	if err != nil {
		return fmt.Errorf("failed to init db: %w", err)
	}
	defer conn.Close()

	nodes, err := db.GetNodeExercisesForClass(conn, clusterID, "")
	if err != nil {
		return fmt.Errorf("failed to get exercise distribution: %w", err)
	}

	nameSet := make(map[string]struct{})
	for _, n := range nodes {
		for _, it := range n.Exercises {
			nameSet[it.Name] = struct{}{}
		}
	}
	if len(nameSet) == 0 {
		fmt.Println(messageUtils.InfoMsg("No exercises found in DB for this cluster"))
		return nil
	}

	names := make([]string, 0, len(nameSet))
	for k := range nameSet {
		names = append(names, k)
	}
	sort.Strings(names)
	picked := fuzzy.NewFuzzyFinder(names, false)
	if len(picked) == 0 {
		fmt.Println(messageUtils.InfoMsg("No exercise selected"))
		return nil
	}
	target := picked[0]

	fmt.Printf("%s %s\n", messageUtils.Bold("Exercise:"), messageUtils.Highlight(target))
	fmt.Println(strings.Repeat("-", 69))
	for _, n := range nodes {
		items := make([]db.ExerciseItem, 0)
		for _, it := range n.Exercises {
			if it.Name == target {
				items = append(items, it)
			}
		}
		if len(items) == 0 {
			continue
		}
		fmt.Printf("\n%s %s\n", messageUtils.Bold("Node:"), messageUtils.Highlight(n.NodeURL))
		sort.Slice(items, func(i, j int) bool { return items[i].GroupName < items[j].GroupName })
		for _, it := range items {
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
	}
	fmt.Println()
	return nil
}
