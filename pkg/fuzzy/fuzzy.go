package fuzzy

import (
	"bytes"
	"fmt"

	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/tidwall/gjson"
)

type FuzzyInfoParams struct {
	Cfg          config.GlobalOptions
	Multi        bool
	Method       string
	Key          string
	ExtraInfo    bool
	ContextType  string
	ContextLabel string
}

func FuzzyInfo(params FuzzyInfoParams) error {
	var selected []gjson.Result
	apiData, vals, err := getValuesForFuzzy(params)
	if err != nil {
		return err
	}
	results := NewFuzzyFinder(vals, params.Multi)
	for _, result := range results {
	outer:
		for _, data := range apiData {
			if element := data.Get(params.Key); element.Exists() && element.String() == result {
				selected = append(selected, data)
				break outer
			}
		}
	}

	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, s := range selected {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(s.Raw)
	}
	buf.WriteByte(']')

	toPrint := buf.Bytes()

	if params.Cfg.Raw {
		utils.PrintJson(toPrint)
		return nil
	} else {
		if params.ExtraInfo && params.ContextType != "" && params.ContextLabel != "" {
			utils.PrintKVWithResourceContext(toPrint, params.ContextType, params.ContextLabel)
		} else {
			utils.PrintKV(toPrint)
		}
		return nil
	}
}

func FuzzyInfoIDs(params FuzzyInfoParams) ([]string, error) {
	var selectedIDs []string
	apiData, vals, err := getValuesForFuzzy(params)
	if err != nil {
		return nil, err
	}

	results := NewFuzzyFinder(vals, params.Multi)

	resourceType := getResourceTypeFromMethod(params.Method)
	idField, _, ok := utils.GetIDFieldMapping(resourceType)
	if !ok {
		return nil, fmt.Errorf("could not determine ID field for resource type: %s", resourceType)
	}

	for _, result := range results {
	outer:
		for _, data := range apiData {
			if element := data.Get(params.Key); element.Exists() && element.String() == result {
				if id := data.Get(idField); id.Exists() {
					selectedIDs = append(selectedIDs, id.String())
				}
				break outer
			}
		}
	}

	return selectedIDs, nil
}

func FuzzyInfoIDsWithData(params FuzzyInfoParams) ([]string, []gjson.Result, error) {
	var selectedIDs []string
	var selectedData []gjson.Result
	apiData, vals, err := getValuesForFuzzy(params)
	if err != nil {
		return nil, nil, err
	}

	results := NewFuzzyFinder(vals, params.Multi)

	resourceType := getResourceTypeFromMethod(params.Method)
	idField, _, ok := utils.GetIDFieldMapping(resourceType)
	if !ok {
		return nil, nil, fmt.Errorf("could not determine ID field for resource type: %s", resourceType)
	}

	for _, result := range results {
	outer:
		for _, data := range apiData {
			if element := data.Get(params.Key); element.Exists() && element.String() == result {
				if id := data.Get(idField); id.Exists() {
					selectedIDs = append(selectedIDs, id.String())
					selectedData = append(selectedData, data)
				}
				break outer
			}
		}
	}

	return selectedIDs, selectedData, nil
}

func getResourceTypeFromMethod(method string) string {
	methodToResource := map[string]string{
		"getUsers":      "user",
		"getGroups":     "group",
		"getRoles":      "role",
		"getPrivileges": "privilege",
		"getACL":        "acl-rule",
		"getTemplates":  "template",
		"getProjects":   "project",
		"getComputes":   "compute",
		"getAppliances": "appliance",
		"getPools":      "pool",
		"getNodes":      "node",
		"getImages":     "image",
		"getLinks":      "link",
		"getDrawings":   "drawing",
		"getSnapshots":  "snapshot",
		"getSymbols":    "symbol",
	}

	if resourceType, ok := methodToResource[method]; ok {
		return resourceType
	}

	if len(method) > 3 && method[:3] == "get" {
		resource := method[3:]
		switch resource {
		case "Users":
			return "user"
		case "Groups":
			return "group"
		case "Roles":
			return "role"
		case "Privileges":
			return "privilege"
		case "Templates":
			return "template"
		case "Projects":
			return "project"
		case "Computes":
			return "compute"
		case "Appliances":
			return "appliance"
		case "Pools":
			return "pool"
		case "Nodes":
			return "node"
		case "Images":
			return "image"
		case "Links":
			return "link"
		case "Drawings":
			return "drawing"
		case "Snapshots":
			return "snapshot"
		case "Symbols":
			return "symbol"
		}
	}

	return "unknown"
}

func getValuesForFuzzy(params FuzzyInfoParams) ([]gjson.Result, []string, error) {
	rawData, _, err := utils.CallClient(params.Cfg, params.Method, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	result := gjson.ParseBytes(rawData)
	if !result.IsArray() {
		return nil, nil, fmt.Errorf("expected array response, got %s", result.Type)
	}

	var apiData []gjson.Result
	var results []string

	result.ForEach(func(_, value gjson.Result) bool {
		apiData = append(apiData, value)
		if val := value.Get(params.Key); val.Exists() {
			results = append(results, val.String())
		}
		return true
	})

	return apiData, results, nil
}

func NewFuzzyInfoParams(cfg config.GlobalOptions, method, key string, multi bool) FuzzyInfoParams {
	return FuzzyInfoParams{
		Cfg:    cfg,
		Multi:  multi,
		Method: method,
		Key:    key,
	}
}

func NewFuzzyInfoParamsWithContext(cfg config.GlobalOptions, method, key string, multi bool, contextType, contextLabel string) FuzzyInfoParams {
	return FuzzyInfoParams{
		Cfg:          cfg,
		Multi:        multi,
		Method:       method,
		Key:          key,
		ExtraInfo:    true,
		ContextType:  contextType,
		ContextLabel: contextLabel,
	}
}
