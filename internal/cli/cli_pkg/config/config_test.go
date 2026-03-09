package config

import (
	"context"
	"testing"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/globals"
)

func TestWithGlobalOptions(t *testing.T) {
	ctx := context.Background()
	opts := GlobalOptions{
		Server:       "http://example.com",
		Insecure:     true,
		KeyFile:      "/path/to/keyfile",
		OutputFormat: globals.OutputJSON,
		CommandPath:  "/test/path",
	}

	newCtx := WithGlobalOptions(ctx, opts)
	if newCtx == nil {
		t.Fatal("WithGlobalOptions() returned nil context")
	}

	retrievedOpts, err := GetGlobalOptionsFromContext(newCtx)
	if err != nil {
		t.Fatalf("GetGlobalOptionsFromContext() error = %v", err)
	}

	if retrievedOpts.Server != opts.Server {
		t.Errorf("Server = %v, want %v", retrievedOpts.Server, opts.Server)
	}
	if retrievedOpts.Insecure != opts.Insecure {
		t.Errorf("Insecure = %v, want %v", retrievedOpts.Insecure, opts.Insecure)
	}
	if retrievedOpts.KeyFile != opts.KeyFile {
		t.Errorf("KeyFile = %v, want %v", retrievedOpts.KeyFile, opts.KeyFile)
	}
	if retrievedOpts.OutputFormat != opts.OutputFormat {
		t.Errorf("OutputFormat = %v, want %v", retrievedOpts.OutputFormat, opts.OutputFormat)
	}
	if retrievedOpts.CommandPath != opts.CommandPath {
		t.Errorf("CommandPath = %v, want %v", retrievedOpts.CommandPath, opts.CommandPath)
	}
}

func TestGetGlobalOptionsFromContext(t *testing.T) {
	t.Run("valid context", func(t *testing.T) {
		opts := GlobalOptions{
			Server:   "http://test.com",
			Insecure: false,
		}
		ctx := WithGlobalOptions(context.Background(), opts)

		retrievedOpts, err := GetGlobalOptionsFromContext(ctx)
		if err != nil {
			t.Errorf("GetGlobalOptionsFromContext() error = %v, want nil", err)
		}
		if retrievedOpts.Server != opts.Server {
			t.Errorf("Server = %v, want %v", retrievedOpts.Server, opts.Server)
		}
	})

	t.Run("context without options", func(t *testing.T) {
		ctx := context.Background()

		_, err := GetGlobalOptionsFromContext(ctx)
		if err == nil {
			t.Error("GetGlobalOptionsFromContext() should return error for context without options")
		}
		expected := "GlobalOptions not found in context"
		if err.Error() != expected {
			t.Errorf("Error message = %v, want %v", err.Error(), expected)
		}
	})

	t.Run("context with wrong type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), optsKey, "not GlobalOptions")

		_, err := GetGlobalOptionsFromContext(ctx)
		if err == nil {
			t.Error("GetGlobalOptionsFromContext() should return error for wrong type")
		}
		expected := "GlobalOptions not found in context"
		if err.Error() != expected {
			t.Errorf("Error message = %v, want %v", err.Error(), expected)
		}
	})
}

func TestGlobalOptionsDefaults(t *testing.T) {
	opts := GlobalOptions{}

	if opts.Server != "" {
		t.Errorf("Default Server = %v, want empty string", opts.Server)
	}
	if opts.Insecure != false {
		t.Errorf("Default Insecure = %v, want false", opts.Insecure)
	}
	if opts.KeyFile != "" {
		t.Errorf("Default KeyFile = %v, want empty string", opts.KeyFile)
	}
	if opts.OutputFormat != globals.OutputFormat(0) {
		t.Errorf("Default OutputFormat = %v, want empty", opts.OutputFormat)
	}
	if opts.CommandPath != "" {
		t.Errorf("Default CommandPath = %v, want empty string", opts.CommandPath)
	}
}

func TestGlobalOptionsContextIsolation(t *testing.T) {
	opts1 := GlobalOptions{Server: "server1.com"}
	opts2 := GlobalOptions{Server: "server2.com"}

	ctx1 := WithGlobalOptions(context.Background(), opts1)
	ctx2 := WithGlobalOptions(context.Background(), opts2)

	retrieved1, _ := GetGlobalOptionsFromContext(ctx1)
	retrieved2, _ := GetGlobalOptionsFromContext(ctx2)

	if retrieved1.Server != opts1.Server {
		t.Errorf("Context 1 Server = %v, want %v", retrieved1.Server, opts1.Server)
	}
	if retrieved2.Server != opts2.Server {
		t.Errorf("Context 2 Server = %v, want %v", retrieved2.Server, opts2.Server)
	}
}

func TestGlobalOptionsContextChain(t *testing.T) {
	baseOpts := GlobalOptions{
		Server:   "http://base.com",
		Insecure: false,
	}
	baseCtx := WithGlobalOptions(context.Background(), baseOpts)

	newOpts := GlobalOptions{
		Server:   "http://new.com",
		Insecure: true,
	}
	newCtx := WithGlobalOptions(baseCtx, newOpts)

	retrieved, err := GetGlobalOptionsFromContext(newCtx)
	if err != nil {
		t.Fatalf("GetGlobalOptionsFromContext() error = %v", err)
	}

	if retrieved.Server != newOpts.Server {
		t.Errorf("Server = %v, want %v", retrieved.Server, newOpts.Server)
	}
	if retrieved.Insecure != newOpts.Insecure {
		t.Errorf("Insecure = %v, want %v", retrieved.Insecure, newOpts.Insecure)
	}

	originalRetrieved, err := GetGlobalOptionsFromContext(baseCtx)
	if err != nil {
		t.Fatalf("GetGlobalOptionsFromContext() error = %v", err)
	}

	if originalRetrieved.Server != baseOpts.Server {
		t.Errorf("Original Server = %v, want %v", originalRetrieved.Server, baseOpts.Server)
	}
}

func TestGlobalOptionsKey(t *testing.T) {
	if optsKey == "" {
		t.Error("optsKey should not be empty")
	}
	if string(optsKey) != "globalOptions" {
		t.Errorf("optsKey = %v, want %v", string(optsKey), "globalOptions")
	}
}
