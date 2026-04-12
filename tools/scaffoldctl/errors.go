package main

import (
	"fmt"
	"strings"
)

type Diagnostic struct {
	Path    string
	Field   string
	Message string
}

type Diagnostics struct {
	items []Diagnostic
}

func (d *Diagnostics) Add(path, field, message string) {
	d.items = append(d.items, Diagnostic{
		Path:    path,
		Field:   field,
		Message: message,
	})
}

func (d *Diagnostics) HasErrors() bool {
	return len(d.items) > 0
}

func (d *Diagnostics) Error() string {
	if len(d.items) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("scaffoldctl found spec errors:\n")
	for _, item := range d.items {
		if item.Field != "" {
			fmt.Fprintf(&b, "  - %s: %s: %s\n", item.Path, item.Field, item.Message)
			continue
		}
		fmt.Fprintf(&b, "  - %s: %s\n", item.Path, item.Message)
	}
	return strings.TrimRight(b.String(), "\n")
}
