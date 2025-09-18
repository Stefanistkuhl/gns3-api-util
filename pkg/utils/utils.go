package utils

import (
	"bufio"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stefanistkuhl/gns3util/pkg/api"
	"github.com/stefanistkuhl/gns3util/pkg/api/endpoints"
	"github.com/stefanistkuhl/gns3util/pkg/authentication"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

func GetIDFieldMapping(resourceType string) (string, string, bool) {
	if fields, ok := idElementName[resourceType]; ok {
		return fields[0], fields[1], true
	}
	return "", "", false
}

func CallClient(cfg config.GlobalOptions, cmdName string, args []string, body any) ([]byte, int, error) {
	cmd, ok := commandMap[cmdName]
	if !ok {
		return nil, 0, fmt.Errorf("unknown command: %s", cmdName)
	}

	token := ""
	if cmdName != "userAuthenticate" {
		err := errors.New("")
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

	client := api.NewGNS3Client(settings)
	reqOpts := api.NewRequestOptions(settings).
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
			b, err := json.Marshal(v)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to encode request body: %w", err)
			}
			dataStr = string(b)
		}
		reqOpts = reqOpts.WithData(dataStr)
	}

	respBody, resp, err := client.Do(reqOpts)
	if err != nil {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		return respBody, status, err
	}
	return respBody, resp.StatusCode, nil
}

func ExecuteAndPrint(cfg config.GlobalOptions, cmdName string, args []string) {
	body, status, err := CallClient(cfg, cmdName, args, nil)
	if err != nil {
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "Authentication was unsuccessful") {
			fmt.Printf("%v Authentication failed. Please check your username and password.\n", colorUtils.Error("Error:"))
			return
		}
		fmt.Printf("%v %v\n", colorUtils.Error("Error:"), err)
		return
	}
	if status == 204 {
		fmt.Printf("%v Command '%s' executed successfully (no content returned)\n",
			colorUtils.Success("Success:"), cmdName)
		return
	}
	if len(body) == 0 {
		fmt.Printf("%v Command '%s' executed successfully (empty response)\n",
			colorUtils.Success("Success:"), cmdName)
		return
	}
	if cfg.Raw {
		if cfg.NoColors {
			PrintJsonUgly(body)
		} else {
			PrintJson(body)
		}
	} else {
		PrintKV(body)
	}
}

func PrintJson(body []byte) {
	result := pretty.Pretty(body)
	result = pretty.Color(result, nil)
	fmt.Print(string(result))
}

func PrintJsonUgly(body []byte) {
	result := pretty.Pretty(body)
	fmt.Print(string(result))
}

func PrintKV(body []byte) {
	result := gjson.ParseBytes(body)

	if result.IsArray() {
		if len(result.Array()) == 0 {
			fmt.Println("  No data found")
			return
		}
		result.ForEach(func(_, elem gjson.Result) bool {
			PrintSeperator()
			if elem.IsObject() {
				elem.ForEach(func(key, value gjson.Result) bool {
					fmt.Printf("  %s: %s\n", colorUtils.Highlight(key.String()), value.Raw)
					return true
				})
			} else {
				fmt.Printf("  %s\n", elem.Raw)
			}
			return true
		})
		PrintSeperator()
	} else if result.IsObject() {
		PrintSeperator()
		result.ForEach(func(key, value gjson.Result) bool {
			fmt.Printf("  %s: %s\n", colorUtils.Highlight(key.String()), value.Raw)
			return true
		})
		PrintSeperator()
	}
}

func PrintKVWithContext(body []byte, contextType, contextField, contextLabel string) {
	result := gjson.ParseBytes(body)

	if result.IsArray() {
		if len(result.Array()) == 0 {
			fmt.Println("  No data found")
			return
		}

		if contextType != "" && contextField != "" {
			contextGroups := make(map[string][]gjson.Result)

			result.ForEach(func(_, elem gjson.Result) bool {
				if elem.IsObject() {
					contextValue := elem.Get(contextField)
					if contextValue.Exists() {
						contextKey := contextValue.String()
						contextGroups[contextKey] = append(contextGroups[contextKey], elem)
					} else {
						contextGroups["Unknown"] = append(contextGroups["Unknown"], elem)
					}
				} else {
					contextGroups["Unknown"] = append(contextGroups["Unknown"], elem)
				}
				return true
			})

			// Print grouped results
			for contextKey, items := range contextGroups {
				if contextLabel != "" {
					fmt.Printf("\n%s %s\n", colorUtils.Bold(contextLabel), colorUtils.Highlight(contextKey))
				}
				fmt.Println(strings.Repeat("-", 40))

				for _, elem := range items {
					if elem.IsObject() {
						elem.ForEach(func(key, value gjson.Result) bool {
							fmt.Printf("  %s: %s\n", colorUtils.Highlight(key.String()), value.Raw)
							return true
						})
					} else {
						fmt.Printf("  %s\n", elem.Raw)
					}
					fmt.Println()
				}
			}
		} else {
			result.ForEach(func(_, elem gjson.Result) bool {
				PrintSeperator()
				if elem.IsObject() {
					elem.ForEach(func(key, value gjson.Result) bool {
						fmt.Printf("  %s: %s\n", colorUtils.Highlight(key.String()), value.Raw)
						return true
					})
				} else {
					fmt.Printf("  %s\n", elem.Raw)
				}
				return true
			})
			PrintSeperator()
		}
	} else if result.IsObject() {
		PrintSeperator()
		result.ForEach(func(key, value gjson.Result) bool {
			fmt.Printf("  %s: %s\n", colorUtils.Highlight(key.String()), value.Raw)
			return true
		})
		PrintSeperator()
		fmt.Println(result.Raw)
	}
}

func PrintKVWithResourceContext(body []byte, resourceType, contextLabel string) {
	if fields, ok := idElementName[resourceType]; ok {
		contextField := fields[1]
		PrintKVWithContext(body, resourceType, contextField, contextLabel)
	} else {
		PrintKV(body)
	}
}

func PrintSeperator() {
	fmt.Println(colorUtils.Seperator(strings.Repeat("-", 69)))
}

func ExecuteAndPrintWithBody(cfg config.GlobalOptions, cmdName string, args []string, body any) {
	respBody, status, err := CallClient(cfg, cmdName, args, body)
	if err != nil {
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "Authentication was unsuccessful") {
			fmt.Printf("%v Authentication failed. Please check your username and password.\n", colorUtils.Error("Error:"))
			return
		}
		fmt.Printf("%v %v\n", colorUtils.Error("Error:"), err)
		return
	}
	if status == 204 {
		fmt.Printf("%v Command '%s' executed successfully (no content returned)\n",
			colorUtils.Success("Success:"), cmdName)
		return
	}
	if len(respBody) == 0 {
		fmt.Printf("%v Command '%s' executed successfully (empty response)\n",
			colorUtils.Success("Success:"), cmdName)
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
func ResolveID(cfg config.GlobalOptions, subcommand string, name string, args []string) (string, error) {
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
	client := api.NewGNS3Client(settings)

	ep := endpoints.Endpoints{}
	endpointPath := cmd.Endpoint(ep, args)

	reqOpts := api.NewRequestOptions(settings).
		WithURL(endpointPath).
		WithMethod(api.GET)

	body, resp, err := client.Do(reqOpts)
	if err != nil {
		return "", fmt.Errorf("API error: %w", err)
	}
	defer resp.Body.Close()

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

	return "", fmt.Errorf("failed to resolve the name %s to a valid id", colorUtils.Bold(name))
}

func GetResourceWithContext(cfg config.GlobalOptions, commandName string, resourceIDs []string, contextType, contextLabel string) (map[string][]byte, error) {
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
				contextKey = getContextKey(contextResult, contextType)
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

func PrintResourceWithContext(resourceData map[string][]byte, contextLabel string) {
	for i, contextKey := range getSortedKeys(resourceData) {
		resourceBody := resourceData[contextKey]

		if i > 0 {
			fmt.Println()
		}

		if contextLabel != "" {
			fmt.Printf("\n%s %s\n", colorUtils.Bold(contextLabel), colorUtils.Highlight(contextKey))
			fmt.Println(strings.Repeat("-", 40))
		}

		if len(resourceBody) == 0 {
			fmt.Println("  No data found")
		} else {
			resourceResult := gjson.ParseBytes(resourceBody)
			if resourceResult.IsArray() && len(resourceResult.Array()) == 0 {
				fmt.Println("  No data found")
			} else {
				PrintKV(resourceBody)
			}
		}
	}
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

func getContextKey(result gjson.Result, resourceType string) string {
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

func getSortedKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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

func ValidateUrl(input string) bool {
	_, err := url.ParseRequestURI(input)
	if err != nil {
		return false
	}
	return true
}

func ValidateUrlWithReturn(input string) *url.URL {
	u, err := url.ParseRequestURI(input)
	if err != nil {
		return nil
	}
	return u
}

func ValidateAndTestUrl(input string) bool {
	if !ValidateUrl(input) {
		return false
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(http.MethodGet, input, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

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

func PrintTable[T any](items []T, columns []Column[T]) {
	if len(items) == 0 {
		fmt.Println("No records found.")
		return
	}

	widths := make([]int, len(columns))
	for i, col := range columns {
		widths[i] = len(col.Header)
	}
	for _, it := range items {
		for i, col := range columns {
			val := col.Value(it)
			if len(val) > widths[i] {
				widths[i] = len(val)
			}
		}
	}

	for i, col := range columns {
		fmt.Printf("%-*s  ", widths[i], col.Header)
	}
	fmt.Println()

	for i := range columns {
		fmt.Print(strings.Repeat("-", widths[i]) + "  ")
	}
	fmt.Println()

	for _, it := range items {
		for i, col := range columns {
			fmt.Printf("%-*s  ", widths[i], col.Value(it))
		}
		fmt.Println()
	}
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
