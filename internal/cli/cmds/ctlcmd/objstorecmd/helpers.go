package objstorecmd

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
)

func selectFilestore(
	cluster *pathutils.ClusterEntry,
	fileStoreName string,
) (*pathutils.ServiceEntry, error) {
	var available []pathutils.ServiceEntry
	for _, node := range cluster.Nodes {
		if node.Type == pathutils.TypeClusterFileStore {
			available = append(available, node)
		}
	}

	if len(available) == 0 {
		return nil, fmt.Errorf("no filestore found in cluster")
	}

	if fileStoreName != "" {
		for i := range available {
			if available[i].ID == fileStoreName {
				return &available[i], nil
			}
		}
		return nil, fmt.Errorf("filestore ID %q not found",
			fileStoreName)
	}

	if len(available) > 1 {
		fmt.Println(
			"Multiple filestores detected. " +
				"Specify one with --filestore-id:",
		)
		for _, fs := range available {
			fmt.Printf(" - %s (%s)\n", fs.ID, fs.URL)
		}
		return nil, fmt.Errorf(
			"ambiguous filestore selection",
		)
	}

	return &available[0], nil
}
