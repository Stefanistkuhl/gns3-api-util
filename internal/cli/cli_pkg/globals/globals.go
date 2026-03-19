package globals

type OutputFormat int

const (
	OutputJSON OutputFormat = iota
	OutputJSONColorless
	OutputCollapsed
	OutputYAML
	OutputTOML
	OutputTable
	OutputTableASCII
)

func (o OutputFormat) String() string {
	return [...]string{
		"json",
		"json-colorless",
		"collapsed",
		"yaml",
		"toml",
		"table",
		"table-ascii",
	}[o]
}

func ParseOutputFormat(s string) OutputFormat {
	switch s {
	case "json":
		return OutputJSON
	case "json-colorless":
		return OutputJSONColorless
	case "collapsed":
		return OutputCollapsed
	case "yaml":
		return OutputYAML
	case "toml":
		return OutputTOML
	case "table":
		return OutputTable
	case "plain", "table-ascii":
		return OutputTableASCII
	default:
		return OutputJSON
	}
}
