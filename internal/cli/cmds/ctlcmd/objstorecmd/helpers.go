package objstorecmd

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/globals"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/api"
	"github.com/spf13/cobra"
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

func resolveFilestoreAndTokenForName(cfg *config.GlobalOptions, fileStoreName string) (*pathutils.ServiceEntry, string, error) {
	clusterName := cfg.Cluster
	keyPath, err := pathutils.ResolveKeyFilePath("")
	if err != nil {
		return nil, "", err
	}
	kf, err := pathutils.LoadGNS3KeysFile(keyPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load keys: %w", err)
	}

	var targetCluster *pathutils.ClusterEntry
	for i := range kf.Clusters {
		if kf.Clusters[i].Name == clusterName {
			targetCluster = &kf.Clusters[i]
			break
		}
	}
	if targetCluster == nil {
		return nil, "", fmt.Errorf("cluster %q not found", clusterName)
	}
	cfg.ClusterEntry.CaCert = targetCluster.CaCert

	fs, err := selectFilestore(targetCluster, fileStoreName)
	if err != nil {
		return nil, "", err
	}
	return fs, cfg.ClusterEntry.Master.AccessToken, nil
}

func newFilestoreClient(cfg *config.GlobalOptions, fileStoreName string) (*api.ClientV2, *pathutils.ServiceEntry, error) {
	fs, token, err := resolveFilestoreAndTokenForName(cfg, fileStoreName)
	if err != nil {
		return nil, nil, err
	}

	if token == "" {
		return nil, nil, fmt.Errorf("no access token for filestore cluster access")
	}

	settings := api.NewSettings(
		api.WithBaseURLV2(fs.URL+"/api/v1"),
		api.WithToken(token),
		api.WithVerify(!cfg.Insecure),
		api.WithCA([]byte(cfg.ClusterEntry.CaCert)),
		api.WithHTTP3(true),
	)

	return api.NewClientV2(&settings), fs, nil
}

func decodeJSONRow[T any](value any) (*T, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal row value: %w", err)
	}

	var row T
	if err := json.Unmarshal(raw, &row); err != nil {
		return nil, fmt.Errorf("decode row value: %w", err)
	}
	return &row, nil
}

func decodeJSONRows[T any](value any) ([]T, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal row slice: %w", err)
	}

	var rows []T
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("decode row slice: %w", err)
	}
	return rows, nil
}

func wantsTableOutput(cfg *config.GlobalOptions) bool {
	return cfg.OutputFormat == globals.OutputTable || cfg.OutputFormat == globals.OutputTableASCII
}

func printRowOrObject[T any](cmd *cobra.Command, cfg *config.GlobalOptions, value any) error {
	printer, err := utils.GetPrinter(cfg.OutputFormat.String())
	if err != nil {
		return fmt.Errorf("printer setup failed: %w", err)
	}

	if !wantsTableOutput(cfg) {
		return printer.PrintObj(value, cmd.OutOrStdout())
	}

	row, err := decodeJSONRow[T](value)
	if err != nil {
		return err
	}

	record, err := asTableRecord(row)
	if err != nil {
		return err
	}

	return printer.PrintObj(record, cmd.OutOrStdout())
}

func printRowsOrObject[T any](cmd *cobra.Command, cfg *config.GlobalOptions, value any, emptyMessage string) error {
	printer, err := utils.GetPrinter(cfg.OutputFormat.String())
	if err != nil {
		return fmt.Errorf("printer setup failed: %w", err)
	}

	if !wantsTableOutput(cfg) {
		return printer.PrintObj(value, cmd.OutOrStdout())
	}

	rows, err := decodeJSONRows[T](value)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		if emptyMessage == "" {
			emptyMessage = "No resources found."
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), emptyMessage)
		return nil
	}

	records, err := asTableRecords(rows)
	if err != nil {
		return err
	}

	return printer.PrintObj(records, cmd.OutOrStdout())
}

func asTableRecord(value any) (utils.TableRecord, error) {
	if record, ok := value.(utils.TableRecord); ok {
		return record, nil
	}

	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return nil, fmt.Errorf("nil table record")
	}
	if rv.Kind() != reflect.Pointer && rv.CanAddr() {
		if record, ok := rv.Addr().Interface().(utils.TableRecord); ok {
			return record, nil
		}
	}

	return nil, fmt.Errorf("%T does not implement utils.TableRecord", value)
}

func asTableRecords(rows any) ([]utils.TableRecord, error) {
	rv := reflect.ValueOf(rows)
	if !rv.IsValid() {
		return nil, fmt.Errorf("nil row collection")
	}
	if rv.Kind() != reflect.Slice {
		return nil, fmt.Errorf("%T is not a slice", rows)
	}

	records := make([]utils.TableRecord, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)

		if elem.Kind() != reflect.Pointer && elem.CanAddr() {
			elem = elem.Addr()
		}

		record, ok := elem.Interface().(utils.TableRecord)
		if !ok {
			return nil, fmt.Errorf("%T does not implement utils.TableRecord", rv.Index(i).Interface())
		}

		records = append(records, record)
	}
	return records, nil
}
