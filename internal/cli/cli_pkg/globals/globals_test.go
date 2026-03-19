package globals

import (
	"testing"
)

func TestOutputFormatString(t *testing.T) {
	tests := []struct {
		format   OutputFormat
		expected string
	}{
		{OutputJSON, "json"},
		{OutputJSONColorless, "json-colorless"},
		{OutputCollapsed, "collapsed"},
		{OutputYAML, "yaml"},
		{OutputTOML, "toml"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.format.String()
			if got != tt.expected {
				t.Errorf("OutputFormat(%d).String() = %v, want %v", tt.format, got, tt.expected)
			}
		})
	}
}

func TestParseOutputFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected OutputFormat
	}{
		{"json", OutputJSON},
		{"json-colorless", OutputJSONColorless},
		{"collapsed", OutputCollapsed},
		{"yaml", OutputYAML},
		{"toml", OutputTOML},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseOutputFormat(tt.input)
			if got != tt.expected {
				t.Errorf("ParseOutputFormat(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestOutputFormatConstants(t *testing.T) {
	if OutputJSON != 1 {
		t.Errorf("OutputJSON = %v, want 1", OutputJSON)
	}
	if OutputJSONColorless != 2 {
		t.Errorf("OutputJSONColorless = %v, want 2", OutputJSONColorless)
	}
	if OutputCollapsed != 3 {
		t.Errorf("OutputCollapsed = %v, want 3", OutputCollapsed)
	}
	if OutputYAML != 4 {
		t.Errorf("OutputYAML = %v, want 4", OutputYAML)
	}
	if OutputTOML != 5 {
		t.Errorf("OutputTOML = %v, want 5", OutputTOML)
	}
}

func TestOutputFormatRoundTrip(t *testing.T) {
	formats := []OutputFormat{
		OutputJSON,
		OutputJSONColorless,
		OutputCollapsed,
		OutputYAML,
		OutputTOML,
	}

	for _, format := range formats {
		str := format.String()
		parsed := ParseOutputFormat(str)
		if parsed != format {
			t.Errorf("Round trip failed: %v -> %q -> %v", format, str, parsed)
		}
	}
}

func TestParseOutputFormatEdgeCases(t *testing.T) {
	edgeCases := []struct {
		name     string
		input    string
		expected OutputFormat
	}{}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseOutputFormat(tc.input)
			if got != tc.expected {
				t.Errorf("ParseOutputFormat(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}
