package config

import (
	"context"
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/globals"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
)

type globalOptionsKey string

const optsKey globalOptionsKey = "globalOptions"

type GlobalOptions struct {
	Server       string
	Insecure     bool
	KeyFile      string
	OutputFormat globals.OutputFormat
	CommandPath  string
	Cluster      string
	ClusterEntry *pathutils.ClusterEntry
	User         string
}

func GetGlobalOptionsFromContext(ctx context.Context) (*GlobalOptions, error) {
	opts, ok := ctx.Value(optsKey).(*GlobalOptions)
	if !ok {
		return nil, fmt.Errorf("GlobalOptions not found in context")
	}
	return opts, nil
}

func WithGlobalOptions(ctx context.Context, opts *GlobalOptions) context.Context {
	return context.WithValue(ctx, optsKey, opts)
}
