package config

import (
	"context"
	"fmt"
)

type globalOptionsKey string

const optsKey globalOptionsKey = "globalOptions"

type GlobalOptions struct {
	Server   string
	Insecure bool
	Raw      bool
	NoColors bool
	KeyFile  string
}

func GetGlobalOptionsFromContext(ctx context.Context) (GlobalOptions, error) {
	opts, ok := ctx.Value(optsKey).(GlobalOptions)
	if !ok {
		return GlobalOptions{}, fmt.Errorf("GlobalOptions not found in context")
	}
	return opts, nil
}

func WithGlobalOptions(ctx context.Context, opts GlobalOptions) context.Context {
	return context.WithValue(ctx, optsKey, opts)
}
