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
	if OutputJSON != 0 {
		t.Errorf("OutputJSON = %v, want 0", OutputJSON)
	}
	if OutputJSONColorless != 1 {
		t.Errorf("OutputJSONColorless = %v, want 1", OutputJSONColorless)
	}
	if OutputCollapsed != 2 {
		t.Errorf("OutputCollapsed = %v, want 2", OutputCollapsed)
	}
	if OutputYAML != 3 {
		t.Errorf("OutputYAML = %v, want 3", OutputYAML)
	}
	if OutputTOML != 4 {
		t.Errorf("OutputTOML = %v, want 4", OutputTOML)
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
	}{
		{"empty string defaults to JSON", "", OutputJSON},
		{"unknown format defaults to JSON", "unknown", OutputJSON},
		{"uppercase not supported", "JSON", OutputJSON},
		{"mixed case not supported", "Json", OutputJSON},
		{"whitespace not supported", " json ", OutputJSON},
		{"typo in format", "jsno", OutputJSON},
		{"plain alias for table-ascii", "plain", OutputTableASCII},
		{"table-ascii valid", "table-ascii", OutputTableASCII},
		{"table valid", "table", OutputTable},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseOutputFormat(tc.input)
			if got != tc.expected {
				t.Errorf("ParseOutputFormat(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestOutputFormatStringRoundTrip(t *testing.T) {
	formats := []OutputFormat{
		OutputJSON,
		OutputJSONColorless,
		OutputCollapsed,
		OutputYAML,
		OutputTOML,
		OutputTable,
		OutputTableASCII,
	}

	for _, fmt := range formats {
		str := fmt.String()
		parsed := ParseOutputFormat(str)
		if parsed != fmt {
			t.Errorf("Round trip failed for format %d: String()=%q, ParseOutputFormat()=%d", fmt, str, parsed)
		}
	}
}

func TestOutputTableFormats(t *testing.T) {
	if OutputTable != 5 {
		t.Errorf("OutputTable = %v, want 5", OutputTable)
	}
	if OutputTableASCII != 6 {
		t.Errorf("OutputTableASCII = %v, want 6", OutputTableASCII)
	}

	if OutputTable.String() != "table" {
		t.Errorf("OutputTable.String() = %q, want %q", OutputTable.String(), "table")
	}
	if OutputTableASCII.String() != "table-ascii" {
		t.Errorf("OutputTableASCII.String() = %q, want %q", OutputTableASCII.String(), "table-ascii")
	}
}
