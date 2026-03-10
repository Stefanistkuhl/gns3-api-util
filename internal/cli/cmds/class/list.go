package class

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"strings"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	clusterutils "github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/clusterUtils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/messageUtils"
	"github.com/spf13/cobra"
)

func NewClassLsCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "ls",
		Short: "List all classes and their distribution across cluster nodes",
		Long: `List all classes and show their distribution across cluster nodes, including:
- Class name and description
- Number of groups per node
- Number of users per group
- Node URLs and their assignments`,
		RunE: runListClasses,
	}

	listCmd.Flags().Bool("db-only", false, "Show only classes from database (skip API calls)")
	listCmd.Flags().Bool("api-only", false, "Show only classes from API (skip database)")
	listCmd.Flags().StringP("cluster", "c", "", "Cluster name")

	return listCmd
}

func runListClasses(cmd *cobra.Command, args []string) error {
	cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get global options: %w", err)
	}

	dbOnly, _ := cmd.Flags().GetBool("db-only")
	apiOnly, _ := cmd.Flags().GetBool("api-only")
	clusterName, _ := cmd.Flags().GetString("cluster")

	if dbOnly && apiOnly {
		return fmt.Errorf("cannot specify both --db-only and --api-only")
	}

	clusterID, err := clusterutils.ResolveClusterID(cfg, clusterName, cmd.Context())
	if err != nil {
		return err
	}

	var classes []ClassDistribution

	if !apiOnly {
		dbClasses, err := getClassDistributionFromDB(clusterID)
		if err != nil {
			fmt.Printf("%v Warning: failed to get class distribution from database: %v\n",
				messageUtils.WarningMsg("Warning"), err)
		} else {
			classes = append(classes, dbClasses...)
		}
	}

	if !dbOnly {
		apiClasses, err := getClassDistributionFromAPI()
		if err != nil {
			if len(classes) == 0 {
				return fmt.Errorf("failed to get class distribution from API: %w", err)
			}
			fmt.Printf("%v Warning: failed to get class distribution from API: %v\n",
				messageUtils.WarningMsg("Warning"), err)
		} else {
			classes = append(classes, apiClasses...)
		}
	}

	if len(classes) == 0 {
		fmt.Printf("%v No classes found\n", messageUtils.InfoMsg("No classes found"))
		return nil
	}

	uniqueClasses := make(map[string]ClassDistribution)
	for _, class := range classes {
		if existing, exists := uniqueClasses[class.Name]; exists {
			maps.Copy(existing.Nodes, class.Nodes)
			uniqueClasses[class.Name] = existing
		} else {
			uniqueClasses[class.Name] = class
		}
	}

	var finalClasses []ClassDistribution
	for _, class := range uniqueClasses {
		finalClasses = append(finalClasses, class)
	}

	utils.PrintTable(finalClasses, []utils.Column[ClassDistribution]{
		{
			Header: "Class Name",
			Value: func(c ClassDistribution) string {
				return c.Name
			},
		},
		{
			Header: "Description",
			Value: func(c ClassDistribution) string {
				if c.Description == "" {
					return "N/A"
				}
				return c.Description
			},
		},
		{
			Header: "Nodes",
			Value: func(c ClassDistribution) string {
				return fmt.Sprintf("%d", len(c.Nodes))
			},
		},
		{
			Header: "Groups",
			Value: func(c ClassDistribution) string {
				totalGroups := 0
				for _, node := range c.Nodes {
					totalGroups += node.GroupCount
				}
				return fmt.Sprintf("%d", totalGroups)
			},
		},
		{
			Header: "Users",
			Value: func(c ClassDistribution) string {
				totalUsers := 0
				for _, node := range c.Nodes {
					totalUsers += node.UserCount
				}
				return fmt.Sprintf("%d", totalUsers)
			},
		},
	})

	for _, class := range finalClasses {
		fmt.Printf("\n%s %s\n", messageUtils.Bold("Class:"), messageUtils.Highlight(class.Name))
		fmt.Println(strings.Repeat("-", 69))
		type nodeRow struct{ Node, Groups, Users string }
		rows := make([]nodeRow, 0, len(class.Nodes))
		for nodeURL, nodeInfo := range class.Nodes {
			rows = append(rows, nodeRow{
				Node:   nodeURL,
				Groups: fmt.Sprintf("%d", nodeInfo.GroupCount),
				Users:  fmt.Sprintf("%d", nodeInfo.UserCount),
			})
		}
		type printable struct{ Node, Groups, Users string }
		p := make([]printable, 0, len(rows))
		for _, r := range rows {
			p = append(p, printable(r))
		}
		utils.PrintTable(p, []utils.Column[printable]{
			{Header: "Node", Value: func(x printable) string { return x.Node }},
			{Header: "Groups", Value: func(x printable) string { return x.Groups }},
			{Header: "Users", Value: func(x printable) string { return x.Users }},
		})

		for nodeURL, nodeInfo := range class.Nodes {
			if len(nodeInfo.GroupNames) == 0 {
				continue
			}
			fmt.Printf("\n  %s\n", messageUtils.Bold(nodeURL))
			names := clusterutils.TruncateList(nodeInfo.GroupNames, 80)
			fmt.Printf("    %s %d\n", messageUtils.Highlight("Group Names:"), len(nodeInfo.GroupNames))
			for _, name := range names {
				fmt.Printf("    - %s\n", name)
			}
		}

		fmt.Println()
	}

	return nil
}

type ClassDistribution struct {
	Name        string
	Description string
	Nodes       map[string]NodeInfo
}

type NodeInfo struct {
	GroupCount int
	UserCount  int
	GroupNames []string
}

func getClassDistributionFromDB(clusterID int) ([]ClassDistribution, error) {
	store, err := db.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to init db: %w", err)
	}

	rawDist, err := store.GetClassDisribution(context.Background(), int64(clusterID))
	if err != nil {
		return nil, fmt.Errorf("failed to get class distribution: %w", err)
	}

	classMap := make(map[string]*ClassDistribution)
	seenGroups := make(map[string]bool)

	for _, row := range rawDist {
		if _, ok := classMap[row.ClassName]; !ok {
			classMap[row.ClassName] = &ClassDistribution{
				Name:        row.ClassName,
				Description: row.ClassDesc.String,
				Nodes:       make(map[string]NodeInfo),
			}
		}
		nodeURL := reflect.ValueOf(row.NodeUrl).String()

		c := classMap[row.ClassName]
		node := c.Nodes[nodeURL]

		node.UserCount++

		groupKey := nodeURL + row.GroupName
		if !seenGroups[groupKey] {
			node.GroupCount++
			node.GroupNames = append(node.GroupNames, row.GroupName)
			seenGroups[groupKey] = true
		}

		c.Nodes[nodeURL] = node
	}

	var result []ClassDistribution
	for _, dist := range classMap {
		result = append(result, *dist)
	}

	return result, nil
}

func getClassDistributionFromAPI() ([]ClassDistribution, error) {
	return []ClassDistribution{}, nil
}
