package utils

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss/table"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/authentication"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/messageUtils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/api"
	"github.com/0xveya/gns3util/pkg/api/endpoints"
	"github.com/0xveya/gns3util/pkg/utils/nwutils"
)

var idElementName = map[string][2]string{
	"user":      {"user_id", "username"},
	"group":     {"user_group_id", "name"},
	"role":      {"role_id", "name"},
	"privilege": {"privilege_id", "name"},
	"acl-rule":  {"ace_id", "path"},
	"template":  {"template_id", "name"},
	"project":   {"project_id", "name"},
	"compute":   {"compute_id", "name"},
	"appliance": {"appliance_id", "name"},
	"pool":      {"resource_pool_id", "name"},
	"node":      {"node_id", "name"},
	"image":     {"path", "path"},
	"link":      {"link_id", "name"},
	"drawing":   {"drawing_id", "name"},
	"snapshot":  {"snapshot_id", "name"},
	"symbol":    {"symbol_id", "filename"},
}

var subcommandKeyMap = map[string]string{
	"user":      "users",
	"group":     "groups",
	"role":      "roles",
	"privilege": "privileges",
	"acl-rule":  "acl",
	"template":  "templates",
	"project":   "projects",
	"compute":   "computes",
	"appliance": "appliances",
	"pool":      "pools",
	"node":      "nodes",
	"symbol":    "symbols",
}

func GetIDFieldMapping(resourceType string) (idField, nameField string, ok bool) {
	if fields, ok := idElementName[resourceType]; ok {
		return fields[0], fields[1], true
	}
	return "", "", false
}

func CallClient(cfg *config.GlobalOptions, cmdName string, args []string, body any) (respBody []byte, statusCode int, err error) {
	cmd, ok := commandMap[cmdName]
	if !ok {
		return nil, 0, fmt.Errorf("unknown command: %s", cmdName)
	}

	token := ""
	if cmdName != "userAuthenticate" {
		token, err = authentication.GetKeyForServer(cfg)
		if err != nil {
			return nil, 0, err
		}
	}

	settings := api.NewSettings(
		api.WithBaseURL(cfg.Server),
		api.WithVerify(!cfg.Insecure),
		api.WithToken(token),
	)

	ep := endpoints.Endpoints{}
	endpointPath := cmd.Endpoint(ep, args)
	if endpointPath == "" {
		return nil, 0, fmt.Errorf("missing required arguments for command: %s", cmdName)
	}

	client := api.NewGNS3Client(&settings)
	reqOpts := api.NewRequestOptions(&settings).
		WithURL(endpointPath).
		WithMethod(cmd.Method)

	if body != nil {
		var dataStr string
		switch v := body.(type) {
		case string:
			dataStr = v
		case []byte:
			dataStr = string(v)
		default:
			var b []byte
			b, err = json.Marshal(v)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to encode request body: %w", err)
			}
			dataStr = string(b)
		}
		reqOpts = reqOpts.WithData(dataStr)
	}

	respBody, resp, err := client.Do(reqOpts)
	if err != nil {
		if resp != nil {
			return respBody, resp.StatusCode, err
		}
		return respBody, 0, err
	}
	defer resp.Body.Close()

	return respBody, resp.StatusCode, nil
}

func ExecuteAndPrint(cfg *config.GlobalOptions, cmdName string, args []string) {
	body, status, err := CallClient(cfg, cmdName, args, nil)
	if err != nil {
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "Authentication was unsuccessful") {
			fmt.Printf("%v Authentication failed. Please check your username and password.\n", messageUtils.ErrorMsg("Authentication failed"))
			return
		}
		fmt.Printf("%v %v\n", messageUtils.ErrorMsg("API error"), err)
		return
	}
	if status == 204 {
		fmt.Printf("%v Command '%s' executed successfully (no content returned)\n",
			messageUtils.SuccessMsg("Command executed successfully"), cmdName)
		return
	}
	if len(body) == 0 {
		fmt.Printf("%v Command '%s' executed successfully (empty response)\n",
			messageUtils.SuccessMsg("Command executed successfully"), cmdName)
		return
	}
	PrintOutput(body, cfg)
}

func ExecuteAndPrintWithBody(cfg *config.GlobalOptions, cmdName string, args []string, body any) {
	respBody, status, err := CallClient(cfg, cmdName, args, body)
	if err != nil {
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "Authentication was unsuccessful") {
			fmt.Printf("%v Authentication failed. Please check your username and password.\n", messageUtils.ErrorMsg("Authentication failed"))
			return
		}
		fmt.Printf("%v %v\n", messageUtils.ErrorMsg("API error"), err)
		return
	}
	if status == 204 {
		fmt.Printf("%v Command '%s' executed successfully (no content returned)\n",
			messageUtils.SuccessMsg("Command executed successfully"), cmdName)
		return
	}
	if len(respBody) == 0 {
		fmt.Printf("%v Command '%s' executed successfully (empty response)\n",
			messageUtils.SuccessMsg("Command executed successfully"), cmdName)
		return
	}
	result := pretty.Pretty(respBody)
	result = pretty.Color(result, nil)
	fmt.Print(string(result))
}

func IsValidUUIDv4(s string) bool {
	u, err := uuid.Parse(s)
	return err == nil && u.Version() == 4
}

func ResolveID(cfg *config.GlobalOptions, subcommand, name string, args []string) (string, error) {
	titleCaser := cases.Title(language.Und)
	key, ok := subcommandKeyMap[subcommand]
	if !ok {
		return "", fmt.Errorf("could not find the method used to resolve this id for subcommand: %s", subcommand)
	}

	cmd, ok := commandMap["get"+titleCaser.String(key)]
	if !ok {
		return "", fmt.Errorf("no command found to fetch list for subcommand: %s", subcommand)
	}

	token, err := authentication.GetKeyForServer(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	settings := api.NewSettings(
		api.WithBaseURL(cfg.Server),
		api.WithVerify(!cfg.Insecure),
		api.WithToken(token),
	)
	client := api.NewGNS3Client(&settings)

	ep := endpoints.Endpoints{}
	endpointPath := cmd.Endpoint(ep, args)

	reqOpts := api.NewRequestOptions(&settings).
		WithURL(endpointPath).
		WithMethod(api.GET)

	body, resp, err := client.Do(reqOpts)
	if err != nil {
		return "", fmt.Errorf("API error: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("failed to close response body: %v", err)
		}
	}()

	var data []map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}

	fields, ok := idElementName[subcommand]
	if !ok {
		return "", fmt.Errorf("no ID/name mapping for subcommand: %s", subcommand)
	}

	idField := fields[0]
	nameField := fields[1]

	for _, entry := range data {
		if entryName, ok := entry[nameField].(string); ok && entryName == name {
			if idVal, ok := entry[idField].(string); ok {
				if _, err := uuid.Parse(idVal); err == nil {
					return idVal, nil
				}
				return idVal, nil
			}
		}
	}

	return "", fmt.Errorf("failed to resolve the name %s to a valid id", messageUtils.Bold(name))
}

func GetResourceWithContext(cfg *config.GlobalOptions, commandName string, resourceIDs []string, contextType, contextLabel string) (map[string][]byte, error) {
	resourceData := make(map[string][]byte)

	needsContext := contextType != "" && contextLabel != ""

	for _, resourceID := range resourceIDs {
		var contextKey string

		if needsContext {
			contextCommand := getContextCommand(contextType)
			if contextCommand != "" {
				contextBody, _, err := CallClient(cfg, contextCommand, []string{resourceID}, nil)
				if err != nil {
					return nil, fmt.Errorf("failed to get %s info for %s: %w", contextType, resourceID, err)
				}

				contextResult := gjson.ParseBytes(contextBody)
				contextKey = getContextKey(&contextResult, contextType)
				if contextKey == "" {
					contextKey = resourceID
				}
			} else {
				contextKey = resourceID
			}
		} else {
			contextKey = resourceID
		}

		resourceBody, _, err := CallClient(cfg, commandName, []string{resourceID}, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get %s for %s: %w", commandName, contextKey, err)
		}

		resourceData[contextKey] = resourceBody
	}

	return resourceData, nil
}

func getContextCommand(resourceType string) string {
	contextCommands := map[string]string{
		"user":      "getUser",
		"group":     "getGroup",
		"role":      "getRole",
		"privilege": "getPrivilege",
		"template":  "getTemplate",
		"project":   "getProject",
		"compute":   "getCompute",
		"appliance": "getAppliance",
		"pool":      "getPool",
		"node":      "getNode",
		"image":     "getImage",
		"link":      "getLink",
		"drawing":   "getDrawing",
		"snapshot":  "getSnapshot",
		"symbol":    "getSymbol",
	}

	return contextCommands[resourceType]
}

func getContextKey(result *gjson.Result, resourceType string) string {
	if fields, ok := idElementName[resourceType]; ok {
		nameField := fields[1]
		return result.Get(nameField).String()
	}

	commonFields := []string{"name", "username", "title", "label"}
	for _, field := range commonFields {
		if value := result.Get(field); value.Exists() {
			return value.String()
		}
	}

	return ""
}

//go:embed static/*
var staticFiles embed.FS

func GetEmbeddedHTML() []byte {
	data, _ := staticFiles.ReadFile("static/index.html")
	return data
}

func GetEmbeddedCSS() []byte {
	data, _ := staticFiles.ReadFile("static/style.css")
	return data
}

func GetEmbeddedJS() []byte {
	data, _ := staticFiles.ReadFile("static/script.js")
	return data
}

func GetEmbeddedFavicon() []byte {
	data, _ := staticFiles.ReadFile("static/favicon.ico")
	return data
}

func ValidateURL(input string) bool {
	_, err := url.ParseRequestURI(input)
	return err == nil
}

func ValidateURLWithReturn(input string) *url.URL {
	u, err := url.ParseRequestURI(input)
	if err != nil {
		return nil
	}
	return u
}

func ValidateAndTestURL(ctx context.Context, input string) bool {
	if !ValidateURL(input) {
		return false
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, input, http.NoBody)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusPermanentRedirect {
		loc := resp.Header.Get("Location")
		if loc == "/static/web-ui/bundled" {
			return true
		}
		return true
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 399 {
		return true
	}

	return false
}

func Deduplicate(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

type Column[T any] struct {
	Header string
	Value  func(item T) string
}

func PrintTableFromBody(body []byte) {
	result := gjson.ParseBytes(body)

	if result.IsObject() {
		var keys []string
		var lastVal gjson.Result
		result.ForEach(func(k, v gjson.Result) bool {
			keys = append(keys, k.String())
			lastVal = v
			return true
		})
		if len(keys) == 1 && lastVal.IsArray() {
			result = lastVal
		}
	}

	var rawData any
	if err := json.Unmarshal([]byte(result.Raw), &rawData); err != nil {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  Error parsing JSON for table output"))
		return
	}

	var rowsMap []map[string]any

	switch v := rawData.(type) {
	case []any:
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				rowsMap = append(rowsMap, flattenMap(m, ""))
			} else {
				rowsMap = append(rowsMap, map[string]any{"VALUE": item})
			}
		}
	case map[string]any:
		rowsMap = append(rowsMap, flattenMap(v, ""))
	default:
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  No data found or unsupported format"))
		return
	}

	if len(rowsMap) == 0 {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  No data found"))
		return
	}

	headerSet := make(map[string]struct{})
	for _, r := range rowsMap {
		for k := range r {
			headerSet[k] = struct{}{}
		}
	}

	var headers []string
	for k := range headerSet {
		headers = append(headers, k)
	}
	sort.Strings(headers)

	var rows [][]string
	for _, r := range rowsMap {
		var rowData []string
		for _, h := range headers {
			val := ""
			if v, exists := r[h]; exists {
				val = fmt.Sprintf("%v", v)
			}
			rowData = append(rowData, val)
		}
		rows = append(rows, rowData)
	}

	var displayHeaders []string
	for _, h := range headers {
		displayHeaders = append(displayHeaders, strings.ToUpper(h))
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		Headers(displayHeaders...).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == -1 {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("12")).
					Bold(true).
					Padding(0, 1)
			}
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Padding(0, 1)
		})

	fmt.Println(t.Render())
}

func flattenMap(m map[string]any, prefix string) map[string]any {
	out := make(map[string]any)

	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch child := v.(type) {
		case map[string]any:
			nested := flattenMap(child, key)
			maps.Copy(out, nested)
		case []any:
			b, _ := json.Marshal(child)
			out[key] = string(b)
		default:
			out[key] = child
		}
	}
	return out
}

func PrintTable[T any](items []T, columns []Column[T]) {
	if len(items) == 0 {
		fmt.Println(lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("  No records found."))
		return
	}

	var headers []string
	for _, col := range columns {
		headers = append(headers, col.Header)
	}

	var rows [][]string
	for _, it := range items {
		var row []string
		for _, col := range columns {
			row = append(row, col.Value(it))
		}
		rows = append(rows, row)
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		Headers(headers...).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("12")).
					Bold(true).
					Padding(0, 1)
			}
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Padding(0, 1)
		})

	fmt.Println(t.Render())
}

func ConfirmPrompt(msg string, defaultYes bool) bool {
	reader := bufio.NewReader(os.Stdin)
	var options string

	if defaultYes {
		options = "[Y/n]"
	} else {
		options = "[y/N]"
	}

	fmt.Printf("%s %s ", msg, options)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	case "":
		return defaultYes
	default:
		return defaultYes
	}
}

func GetUserInKeyFileForURL(cfg *config.GlobalOptions) (string, error) {
	keyFileLocation, err := pathutils.ResolveKeyFilePath(cfg.KeyFile)
	if err != nil {
		return "", err
	}

	kf, err := pathutils.LoadGNS3KeysFile(keyFileLocation)
	if err != nil {
		return "", fmt.Errorf("failed to load keys: %w", err)
	}

	for _, entry := range kf.StandaloneGNS3 {
		if nwutils.NormalizeURL(entry.URL) == nwutils.NormalizeURL(cfg.Server) {
			return entry.User, nil
		}
	}

	for i := range kf.Clusters {
		cluster := &kf.Clusters[i]
		for j := range cluster.GNS3Servers {
			entry := &cluster.GNS3Servers[j]
			if nwutils.NormalizeURL(entry.URL) == nwutils.NormalizeURL(cfg.Server) {
				return entry.User, nil
			}
		}
	}

	return "", fmt.Errorf("failed to get a matching entry in key file for the url %s", cfg.Server)
}
