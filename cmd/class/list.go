package class

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/cluster/db"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
)

func NewClassLsCmd() *cobra.Command {
	var listCmd = &cobra.Command{
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

	clusterID, err := resolveClusterID(cfg, clusterName)
	if err != nil {
		return err
	}

	var classes []ClassDistribution

	if !apiOnly {
		dbClasses, err := getClassDistributionFromDB(cfg, clusterID)
		if err != nil {
			fmt.Printf("%v Warning: failed to get class distribution from database: %v\n",
				messageUtils.WarningMsg("Warning"), err)
		} else {
			classes = append(classes, dbClasses...)
		}
	}

	if !dbOnly {
		apiClasses, err := getClassDistributionFromAPI(cfg)
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
			for nodeURL, nodeInfo := range class.Nodes {
				existing.Nodes[nodeURL] = nodeInfo
			}
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
			names := truncateList(nodeInfo.GroupNames, 80)
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

func getClassDistributionFromDB(cfg config.GlobalOptions, clusterID int) ([]ClassDistribution, error) {
	conn, err := db.InitIfNeeded()
	if err != nil {
		return nil, fmt.Errorf("failed to init db: %w", err)
	}
	defer conn.Close()

	rows, qerr := conn.Query(`
SELECT
  c.name AS class_name,
  c.description AS class_desc,
  n.protocol || '://' || n.host || ':' || CAST(n.port AS TEXT) AS node_url,
  g.name AS group_name,
  u.username AS username
FROM classes c
JOIN groups g
  ON g.class_id = c.class_id
JOIN users u
  ON u.group_id = g.group_id
JOIN group_assignments ga
  ON ga.group_id = g.group_id
JOIN nodes n
  ON n.node_id = ga.node_id
WHERE c.cluster_id = ?
ORDER BY c.name, n.node_id, g.group_id, u.user_id;
`, clusterID)
	if qerr != nil {
		return nil, fmt.Errorf("failed to query class distribution: %w", qerr)
	}
	defer rows.Close()

	classMap := make(map[string]ClassDistribution)
	for rows.Next() {
		var className, classDesc, nodeURL, groupName, username string
		if err := rows.Scan(&className, &classDesc, &nodeURL, &groupName, &username); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		dist, ok := classMap[className]
		if !ok {
			dist = ClassDistribution{
				Name:        className,
				Description: classDesc,
				Nodes:       make(map[string]NodeInfo),
			}
		}
		node := dist.Nodes[nodeURL]
		if node.GroupNames == nil {
			node.GroupNames = []string{}
		}
		seen := false
		for _, gn := range node.GroupNames {
			if gn == groupName {
				seen = true
				break
			}
		}
		if !seen {
			node.GroupNames = append(node.GroupNames, groupName)
			node.GroupCount++
		}
		node.UserCount++
		dist.Nodes[nodeURL] = node
		classMap[className] = dist
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iter rows: %w", err)
	}

	var result []ClassDistribution
	for _, v := range classMap {
		result = append(result, v)
	}
	return result, nil
}

func getClassDistributionFromAPI(cfg config.GlobalOptions) ([]ClassDistribution, error) {
	return []ClassDistribution{}, nil
}

func resolveClusterID(cfg config.GlobalOptions, clusterName string) (int, error) {
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
			if c.Name == clusterName {
				return c.Id, nil
			}
		}
		return 0, fmt.Errorf("cluster not found: %s", clusterName)
	}

	if cfg.Server == "" {
		return 0, fmt.Errorf("no server configured; use -s or provide -c cluster name")
	}
	urlObj := utils.ValidateUrlWithReturn(cfg.Server)
	if urlObj == nil {
		return 0, fmt.Errorf("invalid server url: %s", cfg.Server)
	}
	derived := fmt.Sprintf("%s%s", urlObj.Hostname(), "_single_node_cluster")
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

func truncateList(items []string, maxLen int) []string {
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
