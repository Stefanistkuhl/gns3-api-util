package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// -----------------------------------------------------------------------------
// QUARANTINE ZONE: LEGACY COMMAND SUPPORT
// -----------------------------------------------------------------------------

func PrintResourceWithContext(resourceData map[string][]byte, contextLabel string) {
	for i, contextKey := range getSortedKeys(resourceData) {
		resourceBody := resourceData[contextKey]

		if i > 0 {
			fmt.Println()
		}

		if contextLabel != "" {
			fmt.Printf("\n\033[1m%s\033[0m \033[36m%s\033[0m\n", contextLabel, contextKey)
			fmt.Println(strings.Repeat("-", 60))
		}

		if len(resourceBody) == 0 || string(resourceBody) == "[]" {
			fmt.Println("  No data found")
			continue
		}

		var data any
		if err := json.Unmarshal(resourceBody, &data); err == nil {
			enc := yaml.NewEncoder(os.Stdout)
			enc.SetIndent(2)
			_ = enc.Encode(data)
			_ = enc.Close()
		} else {
			fmt.Println(string(resourceBody))
		}
	}
}

func getSortedKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
