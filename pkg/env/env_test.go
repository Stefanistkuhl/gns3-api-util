package env

import (
	"os"
	"reflect"
	"testing"
)

type TestConfig struct {
	StringField string `env:"STRING_FIELD" type:"string"`
	IntField    int    `env:"INT_FIELD" type:"int"`
	PortField   int    `env:"PORT_FIELD" type:"port"`
	BoolField   bool   `env:"BOOL_FIELD" type:"bool"`
	URLField    string `env:"URL_FIELD" type:"url"`
	ListenField string `env:"LISTEN_FIELD" type:"listen"`
	SecretField string `env:"SECRET_FIELD" type:"secret"`
	Required    string `env:"REQUIRED" type:"string" required:"true"`
	Optional    string `env:"OPTIONAL" type:"string"`
	WithDefault string `env:"WITH_DEFAULT" type:"string" default:"default_value"`
	NoTag       string
	InvalidTag  string `env:"invalid"`
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    bool
		wantErr bool
	}{
		{"true lower", "true", true, false},
		{"TRUE upper", "TRUE", true, false},
		{"1", "1", true, false},
		{"yes", "yes", true, false},
		{"on", "on", true, false},
		{"false lower", "false", false, false},
		{"FALSE upper", "FALSE", false, false},
		{"0", "0", false, false},
		{"no", "no", false, false},
		{"off", "off", false, false},
		{"invalid", "maybe", false, true},
		{"empty", "", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBool(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBool(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidBool(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid true", "true", true},
		{"valid false", "false", true},
		{"valid 1", "1", true},
		{"valid 0", "0", true},
		{"valid yes", "yes", true},
		{"valid no", "no", true},
		{"valid on", "on", true},
		{"valid off", "off", true},
		{"invalid", "maybe", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidBool(tt.input)
			if got != tt.want {
				t.Errorf("isValidBool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseFieldTag(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		tag      reflect.StructTag
		expected *FieldConfig
	}{
		{
			name:   "simple string field",
			envVar: "TEST_FIELD",
			tag:    `env:"TEST_FIELD" type:"string"`,
			expected: &FieldConfig{
				EnvVar:   "TEST_FIELD",
				Type:     EnvTypeString,
				Required: false,
				Default:  "",
				Validate: getValidator(EnvTypeString),
			},
		},
		{
			name:   "required port field",
			envVar: "PORT",
			tag:    `env:"PORT" type:"port" required:"true"`,
			expected: &FieldConfig{
				EnvVar:   "PORT",
				Type:     EnvTypePort,
				Required: true,
				Default:  "",
				Validate: getValidator(EnvTypePort),
			},
		},
		{
			name:   "field with default",
			envVar: "TIMEOUT",
			tag:    `env:"TIMEOUT" type:"int" default:"30"`,
			expected: &FieldConfig{
				EnvVar:   "TIMEOUT",
				Type:     EnvTypeInt,
				Required: false,
				Default:  "30",
				Validate: getValidator(EnvTypeInt),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFieldTag(tt.envVar, tt.tag)
			if got == nil {
				t.Fatal("parseFieldTag() returned nil")
			}

			if got.EnvVar != tt.expected.EnvVar {
				t.Errorf("EnvVar = %v, want %v", got.EnvVar, tt.expected.EnvVar)
			}
			if got.Type != tt.expected.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.expected.Type)
			}
			if got.Required != tt.expected.Required {
				t.Errorf("Required = %v, want %v", got.Required, tt.expected.Required)
			}
			if got.Default != tt.expected.Default {
				t.Errorf("Default = %v, want %v", got.Default, tt.expected.Default)
			}
		})
	}
}

func TestGetValidator(t *testing.T) {
	tests := []struct {
		name     string
		envType  EnvType
		input    string
		expected bool
	}{
		{"port valid", EnvTypePort, "8080", true},
		{"port invalid", EnvTypePort, "99999", false},
		{"port empty", EnvTypePort, "", false},
		{"listen valid IP", EnvTypeListen, "192.168.1.1", true},
		{"listen valid hostname", EnvTypeListen, "example.com", true},
		{"listen invalid", EnvTypeListen, "", false},
		{"url valid", EnvTypeURL, "http://example.com", true},
		{"url invalid", EnvTypeURL, "not-a-url", false},
		{"secret valid", EnvTypeSecret, "abcdefgh", true},
		{"secret too short", EnvTypeSecret, "abc", false},
		{"bool valid", EnvTypeBool, "true", true},
		{"bool invalid", EnvTypeBool, "maybe", false},
		{"string non-empty", EnvTypeString, "hello", true},
		{"string empty", EnvTypeString, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := getValidator(tt.envType)
			got := validator(tt.input)
			if got != tt.expected {
				t.Errorf("getValidator(%v)(%q) = %v, want %v", tt.envType, tt.input, got, tt.expected)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Store original env vars
	originalVals := make(map[string]string)
	envVars := []string{"STRING_FIELD", "INT_FIELD", "PORT_FIELD", "BOOL_FIELD", "URL_FIELD", "REQUIRED"}
	for _, key := range envVars {
		originalVals[key] = os.Getenv(key)
	}

	// Clean up after test
	defer func() {
		for key, val := range originalVals {
			if val == "" {
				if err := os.Unsetenv(key); err != nil {
					t.Logf("Failed to unset env var %s: %v", key, err)
				}
			} else {
				if err := os.Setenv(key, val); err != nil {
					t.Logf("Failed to set env var %s: %v", key, err)
				}
			}
		}
	}()

	t.Run("successful config load", func(t *testing.T) {
		if err := os.Setenv("STRING_FIELD", "test_value"); err != nil {
			t.Logf("Failed to set STRING_FIELD: %v", err)
		}
		if err := os.Setenv("INT_FIELD", "42"); err != nil {
			t.Logf("Failed to set INT_FIELD: %v", err)
		}
		if err := os.Setenv("PORT_FIELD", "8080"); err != nil {
			t.Logf("Failed to set PORT_FIELD: %v", err)
		}
		if err := os.Setenv("BOOL_FIELD", "true"); err != nil {
			t.Logf("Failed to set BOOL_FIELD: %v", err)
		}
		if err := os.Setenv("URL_FIELD", "http://example.com"); err != nil {
			t.Logf("Failed to set URL_FIELD: %v", err)
		}
		if err := os.Setenv("REQUIRED", "required_value"); err != nil {
			t.Logf("Failed to set REQUIRED: %v", err)
		}

		var config TestConfig
		err := LoadConfig(&config)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		if config.StringField != "test_value" {
			t.Errorf("StringField = %v, want %v", config.StringField, "test_value")
		}
		if config.IntField != 42 {
			t.Errorf("IntField = %v, want %v", config.IntField, 42)
		}
		if config.PortField != 8080 {
			t.Errorf("PortField = %v, want %v", config.PortField, 8080)
		}
		if !config.BoolField {
			t.Errorf("BoolField = %v, want %v", config.BoolField, true)
		}
		if config.URLField != "http://example.com" {
			t.Errorf("URLField = %v, want %v", config.URLField, "http://example.com")
		}
		if config.Required != "required_value" {
			t.Errorf("Required = %v, want %v", config.Required, "required_value")
		}
	})

	t.Run("missing required field", func(t *testing.T) {
		// Clear required field
		if err := os.Unsetenv("REQUIRED"); err != nil {
			t.Logf("Failed to unset REQUIRED: %v", err)
		}

		var config TestConfig
		err := LoadConfig(&config)
		if err == nil {
			t.Error("LoadConfig() expected error for missing required field")
		}
	})

	t.Run("default value", func(t *testing.T) {
		// Set required field and don't set WITH_DEFAULT, should use default
		if err := os.Setenv("REQUIRED", "test_required"); err != nil {
			t.Logf("Failed to set REQUIRED: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("REQUIRED"); err != nil {
				t.Logf("Failed to unset REQUIRED: %v", err)
			}
		}()

		var config TestConfig
		err := LoadConfig(&config)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		if config.WithDefault != "default_value" {
			t.Errorf("WithDefault = %v, want %v", config.WithDefault, "default_value")
		}
	})

	t.Run("invalid input type", func(t *testing.T) {
		var config int
		err := LoadConfig(config)
		if err == nil {
			t.Error("LoadConfig() expected error for non-struct input")
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		err := LoadConfig(nil)
		if err == nil {
			t.Error("LoadConfig() expected error for nil pointer")
		}
	})
}

func TestSetFieldValue(t *testing.T) {
	tests := []struct {
		name     string
		fieldVal reflect.Value
		envVal   string
		cfg      *FieldConfig
		wantErr  bool
	}{
		{
			name:     "string field",
			fieldVal: reflect.ValueOf(new(string)).Elem(),
			envVal:   "test",
			cfg:      &FieldConfig{Type: EnvTypeString},
			wantErr:  false,
		},
		{
			name:     "int field",
			fieldVal: reflect.ValueOf(new(int)).Elem(),
			envVal:   "8080",
			cfg:      &FieldConfig{Type: EnvTypePort},
			wantErr:  false,
		},
		{
			name:     "bool field",
			fieldVal: reflect.ValueOf(new(bool)).Elem(),
			envVal:   "true",
			cfg:      &FieldConfig{Type: EnvTypeBool},
			wantErr:  false,
		},
		{
			name:     "type mismatch",
			fieldVal: reflect.ValueOf(new(string)).Elem(),
			envVal:   "8080",
			cfg:      &FieldConfig{Type: EnvTypePort},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setFieldValue(tt.fieldVal, tt.envVal, tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("setFieldValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
