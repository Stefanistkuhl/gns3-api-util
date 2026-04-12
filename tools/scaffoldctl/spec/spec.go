package spec

type SpecFile struct {
	Path          string
	Version       int                `yaml:"version"`
	Package       PackageMeta        `yaml:"package"`
	Client        ClientConfig       `yaml:"client"`
	ClientMethods []ClientMethodSpec `yaml:"client_methods"`
	CLI           CLIConfig          `yaml:"cli,omitempty"`
	CLICommands   []CLICommandSpec   `yaml:"cli_commands,omitempty"`
}

type PackageMeta struct {
	ID      string        `yaml:"id"`
	Title   string        `yaml:"title,omitempty"`
	Service string        `yaml:"service"`
	Source  PackageSource `yaml:"source,omitempty"`
}

type PackageSource struct {
	VCS      string `yaml:"vcs,omitempty"`
	CommitID string `yaml:"commit_id,omitempty"`
	ChangeID string `yaml:"change_id,omitempty"`
	Summary  string `yaml:"summary,omitempty"`
}

type ClientConfig struct {
	TargetFile     string `yaml:"target_file"`
	DefaultSection string `yaml:"default_section"`
	EndpointGroup  string `yaml:"endpoint_group,omitempty"`
}

type ClientMethodSpec struct {
	ID         string          `yaml:"id"`
	Section    string          `yaml:"section,omitempty"`
	Method     string          `yaml:"method"`
	HTTPMethod string          `yaml:"http_method"`
	URLExpr    string          `yaml:"url_expr"`
	Args       []MethodArgSpec `yaml:"args,omitempty"`
	Request    *RequestSpec    `yaml:"request,omitempty"`
	Response   ResponseSpec    `yaml:"response"`
}

type MethodArgSpec struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type RequestSpec struct {
	Body    bool   `yaml:"body"`
	ArgName string `yaml:"arg_name,omitempty"`
	Type    string `yaml:"type,omitempty"`
}

type ResponseSpec struct {
	PointerType string `yaml:"pointer_type"`
	ValueType   string `yaml:"value_type"`
	DecodeVar   string `yaml:"decode_var"`
	DecodeError string `yaml:"decode_error"`
	NoResponse  bool   `yaml:"no_response,omitempty"`
}

type CLIConfig struct {
	TargetFile     string         `yaml:"target_file,omitempty"`
	ParentFile     string         `yaml:"parent_file,omitempty"`
	ParentFactory  string         `yaml:"parent_factory,omitempty"`
	CommandFactory string         `yaml:"command_factory,omitempty"`
	Use            string         `yaml:"use,omitempty"`
	Aliases        []string       `yaml:"aliases,omitempty"`
	Short          string         `yaml:"short,omitempty"`
	Long           string         `yaml:"long,omitempty"`
	Example        string         `yaml:"example,omitempty"`
	AuthMode       string         `yaml:"auth_mode,omitempty"`
	GroupID        string         `yaml:"group_id,omitempty"`
	Groups         []CLIGroupSpec `yaml:"groups,omitempty"`
	Rows           []CLIRowSpec   `yaml:"rows,omitempty"`
}

type CLIGroupSpec struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
}

type CLIRowSpec struct {
	Type   string            `yaml:"type"`
	Fields []CLIRowFieldSpec `yaml:"fields"`
}

type CLIRowFieldSpec struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type,omitempty"`
	JSON   string `yaml:"json,omitempty"`
	Header string `yaml:"header,omitempty"`
	Value  string `yaml:"value,omitempty"`
}

type CLICommandSpec struct {
	ID           string              `yaml:"id"`
	Factory      string              `yaml:"factory,omitempty"`
	Use          string              `yaml:"use"`
	Aliases      []string            `yaml:"aliases,omitempty"`
	Short        string              `yaml:"short"`
	Long         string              `yaml:"long,omitempty"`
	Example      string              `yaml:"example,omitempty"`
	AuthMode     string              `yaml:"auth_mode,omitempty"`
	GroupID      string              `yaml:"group_id,omitempty"`
	CobraArgs    string              `yaml:"cobra_args,omitempty"`
	ClientMethod string              `yaml:"client_method,omitempty"`
	Helper       string              `yaml:"helper,omitempty"`
	Positionals  []CLIPositionalSpec `yaml:"positionals,omitempty"`
	Flags        []CLIFlagSpec       `yaml:"flags,omitempty"`
	Output       CLIOutputSpec       `yaml:"output,omitempty"`
}

type CLIPositionalSpec struct {
	Name      string `yaml:"name"`
	Usage     string `yaml:"usage,omitempty"`
	ClientArg string `yaml:"client_arg,omitempty"`
}

type CLIFlagSpec struct {
	Name       string `yaml:"name"`
	Shorthand  string `yaml:"shorthand,omitempty"`
	Type       string `yaml:"type"`
	Default    string `yaml:"default,omitempty"`
	Usage      string `yaml:"usage"`
	Required   bool   `yaml:"required,omitempty"`
	ModelField string `yaml:"model_field,omitempty"`
}

type CLIOutputSpec struct {
	ResponseField string `yaml:"response_field,omitempty"`
	RowType       string `yaml:"row_type,omitempty"`
	EmptyMessage  string `yaml:"empty_message,omitempty"`
}
