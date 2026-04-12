package create

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/0xveya/gns3util/tools/scaffoldctl/rgjson"
	"github.com/0xveya/gns3util/tools/scaffoldctl/spec"
)

const defaultCLIAuthMode = "cluster-only"

type CLIFileData struct {
	PackageName   string
	SpecPath      string
	PackageID     string
	Config        CLIConfigData
	Groups        []CLIGroupData
	Rows          []CLIRowData
	Commands      []CLICommandData
	NeedsFmt      bool
	NeedsOS       bool
	NeedsFilepath bool
	NeedsTime     bool
	NeedsConfig   bool
	NeedsModels   bool
}

type CLIConfigData struct {
	spec.CLIConfig
	AuthMode string
}

type CLIGroupData struct {
	spec.CLIGroupSpec
	VarName string
}

type CLICommandData struct {
	spec.CLICommandSpec
	Factory       string
	VarName       string
	CobraArgs     string
	AuthMode      string
	Flags         []CLIFlagData
	Variables     []CLIVariableData
	FlagBindings  []string
	RequiredFlags []string
	RunELines     []string
}

type CLIVariableData struct {
	Name string
	Type string
}

type CLIRowData struct {
	spec.CLIRowSpec
	Receiver string
	Fields   []CLIRowFieldData
}

type CLIRowFieldData struct {
	spec.CLIRowFieldSpec
	GoType   string
	JSONName string
	Header   string
	RowValue string
}

type CLIFlagData struct {
	spec.CLIFlagSpec
	VarName  string
	BindLine string
}

type cliClientMethodMeta struct {
	Spec         spec.ClientMethodSpec
	RequestType  string
	RequestField map[string]rgjson.StructField
}

type cliCommandKind string

const (
	cliCommandKindStub   cliCommandKind = "stub"
	cliCommandKindList   cliCommandKind = "list"
	cliCommandKindGet    cliCommandKind = "get"
	cliCommandKindUpload cliCommandKind = "upload"
	cliCommandKindUpdate cliCommandKind = "update"
	cliCommandKindDelete cliCommandKind = "delete"
)

var nonIdentifierCharsRE = regexp.MustCompile(`[^A-Za-z0-9]+`)

func CreateCLIFile(specFile *spec.SpecFile, commands []spec.CLICommandSpec, templatePath string) (string, error) {
	data, err := NewCLIFileData(specFile, commands)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(filepath.Base(templatePath)).
		Funcs(template.FuncMap{
			"goString":      strconv.Quote,
			"goStringSlice": goStringSlice,
		}).
		ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, filepath.Base(templatePath), data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func NewCLIFileData(specFile *spec.SpecFile, commands []spec.CLICommandSpec) (CLIFileData, error) {
	cfg := specFile.CLI
	authMode := cfg.AuthMode
	if authMode == "" {
		authMode = defaultCLIAuthMode
	}

	groups := make([]CLIGroupData, 0, len(cfg.Groups))
	for _, group := range cfg.Groups {
		groups = append(groups, CLIGroupData{
			CLIGroupSpec: group,
			VarName:      lowerCamel(group.ID) + "Group",
		})
	}

	rows := buildCLIRows(cfg.Rows)
	methodMeta, err := buildClientMethodMeta(specFile.ClientMethods)
	if err != nil {
		return CLIFileData{}, err
	}

	data := CLIFileData{
		PackageName: packageNameFromTarget(cfg.TargetFile),
		SpecPath:    specFile.Path,
		PackageID:   specFile.Package.ID,
		Config: CLIConfigData{
			CLIConfig: cfg,
			AuthMode:  authMode,
		},
		Groups:      groups,
		Rows:        rows,
		NeedsFmt:    true,
		NeedsConfig: false,
		NeedsModels: false,
	}

	for i := range rows {
		row := &rows[i]
		for i := range row.Fields {
			field := &row.Fields[i]
			if strings.Contains(field.GoType, "time.") || strings.Contains(field.RowValue, "time.") {
				data.NeedsTime = true
			}
		}
	}

	commandData := make([]CLICommandData, 0, len(commands))
	for i := range commands {
		command := commands[i]
		meta := methodMeta[command.ClientMethod]
		item, err := buildCLICommandData(specFile, &command, authMode, &meta)
		if err != nil {
			return CLIFileData{}, err
		}
		commandData = append(commandData, item)
		if len(item.RunELines) > 1 {
			data.NeedsConfig = true
		}
		if usesUploadCommand(item.RunELines) {
			data.NeedsOS = true
			data.NeedsFilepath = true
		}
		if commandNeedsModels(item.RunELines) {
			data.NeedsModels = true
		}
	}

	data.Commands = commandData
	return data, nil
}

func buildClientMethodMeta(methods []spec.ClientMethodSpec) (map[string]cliClientMethodMeta, error) {
	out := make(map[string]cliClientMethodMeta, len(methods))
	for i := range methods {
		method := methods[i]
		meta := cliClientMethodMeta{Spec: method}
		if method.Request != nil && method.Request.Type != "" {
			meta.RequestType = strings.TrimSpace(method.Request.Type)
			fields, err := rgjson.FindStructFields(baseTypeName(method.Request.Type))
			if err != nil {
				return nil, fmt.Errorf("find request model for %s: %w", method.Method, err)
			}
			meta.RequestField = fields
		}
		out[method.Method] = meta
	}
	return out, nil
}

func buildCLIRows(rows []spec.CLIRowSpec) []CLIRowData {
	out := make([]CLIRowData, 0, len(rows))
	for _, row := range rows {
		rowData := CLIRowData{
			CLIRowSpec: row,
			Receiver:   "r",
			Fields:     make([]CLIRowFieldData, 0, len(row.Fields)),
		}
		if receiver := lowerCamel(row.Type); receiver != "" {
			rowData.Receiver = receiver[:1]
		}
		for _, field := range row.Fields {
			goType := strings.TrimSpace(field.Type)
			if goType == "" {
				goType = "string"
			}

			jsonName := strings.TrimSpace(field.JSON)
			if jsonName == "" {
				jsonName = snakeCase(field.Name)
			}

			header := strings.TrimSpace(field.Header)
			if header == "" {
				header = strings.ToUpper(strings.Join(identifierParts(field.Name), " "))
			}

			rowValue := strings.TrimSpace(field.Value)
			if rowValue == "" {
				rowValue = rowData.Receiver + "." + field.Name
				if goType != "string" {
					rowValue = "fmt.Sprint(" + rowValue + ")"
				}
			}

			rowData.Fields = append(rowData.Fields, CLIRowFieldData{
				CLIRowFieldSpec: field,
				GoType:          goType,
				JSONName:        jsonName,
				Header:          header,
				RowValue:        rowValue,
			})
		}
		out = append(out, rowData)
	}
	return out
}

func buildCLICommandData(
	specFile *spec.SpecFile,
	command *spec.CLICommandSpec,
	defaultAuthMode string,
	meta *cliClientMethodMeta,
) (CLICommandData, error) {
	commandAuthMode := command.AuthMode
	if commandAuthMode == "" {
		commandAuthMode = defaultAuthMode
	}

	factory := strings.TrimSpace(command.Factory)
	if factory == "" {
		factory = "new" + pascal(command.ID) + "Cmd"
	}

	cobraArgs := strings.TrimSpace(command.CobraArgs)
	if cobraArgs == "" {
		cobraArgs = "cobra.NoArgs"
	}

	flags := buildCLIFlags(command.Flags)
	item := CLICommandData{
		CLICommandSpec: *command,
		Factory:        factory,
		VarName:        lowerCamel(command.ID) + "Cmd",
		CobraArgs:      cobraArgs,
		AuthMode:       commandAuthMode,
		Flags:          flags,
	}

	kind := inferCLICommandKind(specFile, command, meta)
	if kind == cliCommandKindStub {
		item.Variables = flagVariables(flags, false)
		item.FlagBindings = flagBindingLines(flags, false)
		item.RequiredFlags = requiredFlagLines(flags)
		item.RunELines = []string{
			fmt.Sprintf("return fmt.Errorf(%q)", implementationHint(specFile.Path, command.ID, command.ClientMethod)),
		}
		return item, nil
	}

	item.Variables = flagVariables(flags, true)
	item.FlagBindings = flagBindingLines(flags, true)
	item.RequiredFlags = requiredFlagLines(flags)

	runELines, err := renderCommandRunELines(command, meta, flags, kind)
	if err != nil {
		return CLICommandData{}, fmt.Errorf("build CLI command %s: %w", command.ID, err)
	}
	item.RunELines = runELines

	return item, nil
}

func buildCLIFlags(flags []spec.CLIFlagSpec) []CLIFlagData {
	out := make([]CLIFlagData, 0, len(flags))
	seenNames := map[string]int{}
	for _, flag := range flags {
		varNameSource := flag.Name
		if flag.ModelField != "" {
			varNameSource = flag.ModelField
		}
		varName := safeGoIdentifier(lowerCamel(varNameSource))
		if seenNames[varName] > 0 {
			seenNames[varName]++
			varName = fmt.Sprintf("%s%d", varName, seenNames[varName])
		} else {
			seenNames[varName] = 1
		}

		flagData := CLIFlagData{
			CLIFlagSpec: flag,
			VarName:     varName,
		}
		flagData.BindLine = buildFlagBindLine(&flagData)
		out = append(out, flagData)
	}
	return out
}

func buildFlagBindLine(flag *CLIFlagData) string {
	method := map[string]string{
		"string": "StringVar",
		"int64":  "Int64Var",
		"bool":   "BoolVar",
	}[flag.Type]
	if method == "" {
		method = "StringVar"
	}

	if flag.Shorthand != "" {
		method += "P"
		return fmt.Sprintf(
			"cmd.Flags().%s(&%s, %s, %s, %s, %s)",
			method,
			flag.VarName,
			strconv.Quote(flag.Name),
			strconv.Quote(flag.Shorthand),
			flagDefaultLiteral(&flag.CLIFlagSpec),
			strconv.Quote(flag.Usage),
		)
	}

	return fmt.Sprintf(
		"cmd.Flags().%s(&%s, %s, %s, %s)",
		method,
		flag.VarName,
		strconv.Quote(flag.Name),
		flagDefaultLiteral(&flag.CLIFlagSpec),
		strconv.Quote(flag.Usage),
	)
}

func flagDefaultLiteral(flag *spec.CLIFlagSpec) string {
	switch flag.Type {
	case "int64":
		if flag.Default == "" {
			return "0"
		}
		return flag.Default
	case "bool":
		if flag.Default == "" {
			return "false"
		}
		return flag.Default
	default:
		return strconv.Quote(flag.Default)
	}
}

func flagVariables(flags []CLIFlagData, includeFilestore bool) []CLIVariableData {
	out := make([]CLIVariableData, 0, len(flags)+1)
	if includeFilestore {
		out = append(out, CLIVariableData{Name: "fileStoreName", Type: "string"})
	}
	for i := range flags {
		flag := &flags[i]
		out = append(out, CLIVariableData{Name: flag.VarName, Type: flag.Type})
	}
	return out
}

func flagBindingLines(flags []CLIFlagData, includeFilestore bool) []string {
	out := make([]string, 0, len(flags)+1)
	for i := range flags {
		flag := &flags[i]
		out = append(out, flag.BindLine)
	}
	if includeFilestore {
		out = append(out, `cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID")`)
	}
	return out
}

func requiredFlagLines(flags []CLIFlagData) []string {
	out := make([]string, 0, len(flags))
	for i := range flags {
		flag := &flags[i]
		if flag.Required {
			out = append(out,
				fmt.Sprintf("if err := cmd.MarkFlagRequired(%q); err != nil {", flag.Name),
				"\tpanic(err)",
				"}",
			)
		}
	}
	return out
}

func inferCLICommandKind(specFile *spec.SpecFile, command *spec.CLICommandSpec, meta *cliClientMethodMeta) cliCommandKind {
	if specFile.Package.Service != "filestore" || packageNameFromTarget(specFile.CLI.TargetFile) != "objstorecmd" {
		return cliCommandKindStub
	}
	if meta.Spec.Method == "" {
		return cliCommandKindStub
	}

	if meta.Spec.Request != nil && meta.Spec.Request.Body && meta.Spec.Response.PointerType == "*models.InitUploadResponse" {
		return cliCommandKindUpload
	}

	switch meta.Spec.HTTPMethod {
	case "GET":
		if command.Output.ResponseField != "" {
			return cliCommandKindList
		}
		return cliCommandKindGet
	case "PATCH":
		if meta.Spec.Request != nil && meta.Spec.Request.Body {
			return cliCommandKindUpdate
		}
	case "DELETE":
		return cliCommandKindDelete
	}

	return cliCommandKindStub
}

func renderCommandRunELines(command *spec.CLICommandSpec, meta *cliClientMethodMeta, flags []CLIFlagData, kind cliCommandKind) ([]string, error) {
	switch kind {
	case cliCommandKindList:
		return renderListRunELines(command, meta), nil
	case cliCommandKindGet:
		return renderGetRunELines(command, meta), nil
	case cliCommandKindUpload:
		return renderUploadRunELines(command, meta, flags)
	case cliCommandKindUpdate:
		return renderUpdateRunELines(command, meta, flags)
	case cliCommandKindDelete:
		return renderDeleteRunELines(command, meta), nil
	default:
		return []string{fmt.Sprintf("return fmt.Errorf(%q)", implementationHint("", command.ID, command.ClientMethod))}, nil
	}
}

func renderListRunELines(command *spec.CLICommandSpec, meta *cliClientMethodMeta) []string {
	lines := commandPrelude()
	lines = append(lines,
		"client, _, err := newFilestoreClient(cfg, fileStoreName)",
		"if err != nil {",
		"\treturn err",
		"}",
		fmt.Sprintf("resp, err := client.%s(cmd.Context())", meta.Spec.Method),
		"if err != nil {",
		"\treturn err",
		"}",
	)
	rowType := strings.TrimSpace(command.Output.RowType)
	if rowType != "" && command.Output.ResponseField != "" {
		lines = append(lines, fmt.Sprintf("return printRowsOrObject[%s](cmd, cfg, resp.%s, %q)", rowType, command.Output.ResponseField, defaultEmptyMessage(command)))
		return lines
	}
	lines = append(lines, "return printObject(cmd, cfg, resp)")
	return lines
}

func renderGetRunELines(command *spec.CLICommandSpec, meta *cliClientMethodMeta) []string {
	lines := commandPrelude()
	lines = append(lines, positionalAssignments(command.Positionals)...)
	lines = append(lines,
		"client, _, err := newFilestoreClient(cfg, fileStoreName)",
		"if err != nil {",
		"\treturn err",
		"}",
		fmt.Sprintf("resp, err := client.%s(cmd.Context(), %s)", meta.Spec.Method, strings.Join(clientPositionalArgs(command.Positionals), ", ")),
		"if err != nil {",
		"\treturn err",
		"}",
	)
	if rowType := strings.TrimSpace(command.Output.RowType); rowType != "" {
		lines = append(lines, fmt.Sprintf("return printRowOrObject[%s](cmd, cfg, resp)", rowType))
		return lines
	}
	lines = append(lines, "return printObject(cmd, cfg, resp)")
	return lines
}

func renderUploadRunELines(command *spec.CLICommandSpec, meta *cliClientMethodMeta, flags []CLIFlagData) ([]string, error) {
	requestType := strings.TrimPrefix(meta.RequestType, "*")
	if requestType == "" {
		return nil, fmt.Errorf("upload command %s has no request model", command.ID)
	}

	lines := commandPrelude()
	lines = append(lines, positionalAssignments(command.Positionals)...)
	lines = append(lines,
		"client, _, err := newFilestoreClient(cfg, fileStoreName)",
		"if err != nil {",
		"\treturn err",
		"}",
		"file, err := os.Open(filePath)",
		"if err != nil {",
		"\treturn err",
		"}",
		"defer file.Close()",
		"stat, err := file.Stat()",
		"if err != nil {",
		"\treturn err",
		"}",
		fmt.Sprintf("req := &%s{", requestType),
		"\tFilename: filepath.Base(filePath),",
		"\tSizeBytes: stat.Size(),",
	)
	requestLines, err := renderNonPointerRequestFields(flags, meta, false)
	if err != nil {
		return nil, err
	}
	lines = append(lines, requestLines...)
	lines = append(lines, "}")
	if _, ok := meta.RequestField["BucketID"]; ok {
		lines = append(lines,
			"if req.BucketID == \"\" {",
			"\treq.BucketID = models.GlobalBucketID",
			"}",
		)
	}
	lines = append(lines, renderPointerRequestFields(flags, meta, false)...)
	rowType := strings.TrimSpace(command.Output.RowType)
	if rowType == "" {
		rowType = "uploadResult"
	}
	lines = append(lines,
		fmt.Sprintf("initResp, err := client.%s(cmd.Context(), req)", meta.Spec.Method),
		"if err != nil {",
		"\treturn err",
		"}",
		"status, err := client.GetUploadStatus(cmd.Context(), initResp.FileUUID)",
		"if err != nil {",
		"\treturn err",
		"}",
		"if status.Offset >= stat.Size() {",
		fmt.Sprintf("\treturn printRowOrObject[%s](cmd, cfg, &models.FinalizeUploadResponse{", rowType),
		"\t\tFileUUID: initResp.FileUUID,",
		"\t\tStatus:   models.FileStatusAvailable,",
		"\t})",
		"}",
		"if status.Offset > 0 {",
		"\tif _, err := file.Seek(status.Offset, 0); err != nil {",
		"\t\treturn fmt.Errorf(\"failed to seek local file: %w\", err)",
		"\t}",
		"}",
		"resp, err := client.StreamUpload(cmd.Context(), initResp.FileUUID, file, status.Offset)",
		"if err != nil {",
		"\treturn err",
		"}",
	)
	if rowType != "" {
		lines = append(lines, fmt.Sprintf("return printRowOrObject[%s](cmd, cfg, resp)", rowType))
	} else {
		lines = append(lines, "return printObject(cmd, cfg, resp)")
	}
	return lines, nil
}

func renderUpdateRunELines(command *spec.CLICommandSpec, meta *cliClientMethodMeta, flags []CLIFlagData) ([]string, error) {
	requestType := strings.TrimPrefix(meta.RequestType, "*")
	if requestType == "" {
		return nil, fmt.Errorf("update command %s has no request model", command.ID)
	}

	lines := commandPrelude()
	lines = append(lines, positionalAssignments(command.Positionals)...)
	lines = append(lines,
		"client, _, err := newFilestoreClient(cfg, fileStoreName)",
		"if err != nil {",
		"\treturn err",
		"}",
		fmt.Sprintf("req := &%s{}", requestType),
	)
	changedNames := make([]string, 0, len(flags))
	for i := range flags {
		flag := &flags[i]
		changedNames = append(changedNames, fmt.Sprintf("cmd.Flags().Changed(%q)", flag.Name))
	}
	if len(changedNames) > 0 {
		lines = append(lines,
			fmt.Sprintf("if !(%s) {", strings.Join(changedNames, " || ")),
			"\treturn fmt.Errorf(\"no update flags were provided\")",
			"}",
		)
	}
	lines = append(lines, renderPointerRequestFields(flags, meta, true)...)
	lines = append(lines, renderChangedValueAssignments(flags, meta)...)
	lines = append(lines,
		fmt.Sprintf("resp, err := client.%s(cmd.Context(), %s, req)", meta.Spec.Method, strings.Join(clientPositionalArgs(command.Positionals), ", ")),
		"if err != nil {",
		"\treturn err",
		"}",
	)
	if rowType := strings.TrimSpace(command.Output.RowType); rowType != "" {
		lines = append(lines, fmt.Sprintf("return printRowOrObject[%s](cmd, cfg, resp)", rowType))
	} else {
		lines = append(lines, "return printObject(cmd, cfg, resp)")
	}
	return lines, nil
}

func renderDeleteRunELines(command *spec.CLICommandSpec, meta *cliClientMethodMeta) []string {
	lines := commandPrelude()
	lines = append(lines, positionalAssignments(command.Positionals)...)
	lines = append(lines,
		"client, _, err := newFilestoreClient(cfg, fileStoreName)",
		"if err != nil {",
		"\treturn err",
		"}",
		fmt.Sprintf("resp, err := client.%s(cmd.Context(), %s)", meta.Spec.Method, strings.Join(clientPositionalArgs(command.Positionals), ", ")),
		"if err != nil {",
		"\treturn err",
		"}",
	)
	if rowType := strings.TrimSpace(command.Output.RowType); rowType != "" {
		lines = append(lines, fmt.Sprintf("return printRowOrObject[%s](cmd, cfg, resp)", rowType))
	} else {
		lines = append(lines, "return printObject(cmd, cfg, resp)")
	}
	return lines
}

func commandPrelude() []string {
	return []string{
		"cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())",
		"if err != nil {",
		"\treturn fmt.Errorf(\"failed to get global options: %w\", err)",
		"}",
	}
}

func positionalAssignments(positionals []spec.CLIPositionalSpec) []string {
	lines := make([]string, 0, len(positionals))
	for i, positional := range positionals {
		name := positionalVarName(positional, i)
		lines = append(lines, fmt.Sprintf("%s := args[%d]", name, i))
	}
	return lines
}

func positionalVarName(positional spec.CLIPositionalSpec, index int) string {
	name := strings.TrimSpace(positional.ClientArg)
	if name == "" {
		name = strings.TrimSpace(positional.Name)
	}
	if name == "" {
		name = fmt.Sprintf("arg%d", index)
	}
	return safeGoIdentifier(lowerCamel(name))
}

func clientPositionalArgs(positionals []spec.CLIPositionalSpec) []string {
	args := make([]string, 0, len(positionals))
	for i, positional := range positionals {
		args = append(args, positionalVarName(positional, i))
	}
	return args
}

func renderNonPointerRequestFields(flags []CLIFlagData, meta *cliClientMethodMeta, changedOnly bool) ([]string, error) {
	var lines []string
	for i := range flags {
		flag := &flags[i]
		if flag.ModelField == "" {
			continue
		}
		field, ok := meta.RequestField[flag.ModelField]
		if !ok {
			return nil, fmt.Errorf("request model %s has no field %s", meta.RequestType, flag.ModelField)
		}
		if strings.HasPrefix(field.Type, "*") {
			continue
		}
		if changedOnly {
			lines = append(lines,
				fmt.Sprintf("if cmd.Flags().Changed(%q) {", flag.Name),
				fmt.Sprintf("\treq.%s = %s", flag.ModelField, flag.VarName),
				"}",
			)
			continue
		}
		lines = append(lines, fmt.Sprintf("\t%s: %s,", flag.ModelField, flag.VarName))
	}
	return lines, nil
}

func renderPointerRequestFields(flags []CLIFlagData, meta *cliClientMethodMeta, onlyChanged bool) []string {
	lines := make([]string, 0, len(flags)*3)
	for i := range flags {
		flag := &flags[i]
		if flag.ModelField == "" {
			continue
		}
		field, ok := meta.RequestField[flag.ModelField]
		if !ok || !strings.HasPrefix(field.Type, "*") {
			continue
		}
		if onlyChanged {
			lines = append(lines,
				fmt.Sprintf("if cmd.Flags().Changed(%q) {", flag.Name),
				fmt.Sprintf("\treq.%s = &%s", flag.ModelField, flag.VarName),
				"}",
			)
			continue
		}
		lines = append(lines,
			fmt.Sprintf("if cmd.Flags().Changed(%q) {", flag.Name),
			fmt.Sprintf("\treq.%s = &%s", flag.ModelField, flag.VarName),
			"}",
		)
	}
	return lines
}

func renderChangedValueAssignments(flags []CLIFlagData, meta *cliClientMethodMeta) []string {
	var lines []string
	for i := range flags {
		flag := &flags[i]
		if flag.ModelField == "" {
			continue
		}
		field, ok := meta.RequestField[flag.ModelField]
		if !ok || strings.HasPrefix(field.Type, "*") {
			continue
		}
		lines = append(lines,
			fmt.Sprintf("if cmd.Flags().Changed(%q) {", flag.Name),
			fmt.Sprintf("\treq.%s = %s", flag.ModelField, flag.VarName),
			"}",
		)
	}
	return lines
}

func usesUploadCommand(lines []string) bool {
	for _, line := range lines {
		if strings.Contains(line, "os.Open(") || strings.Contains(line, "filepath.Base(") {
			return true
		}
	}
	return false
}

func commandNeedsModels(lines []string) bool {
	for _, line := range lines {
		if strings.Contains(line, "models.") {
			return true
		}
	}
	return false
}

func defaultEmptyMessage(command *spec.CLICommandSpec) string {
	if strings.TrimSpace(command.Output.EmptyMessage) != "" {
		return command.Output.EmptyMessage
	}
	return "No resources found."
}

func baseTypeName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "*")
	s = strings.TrimPrefix(s, "models.")
	return s
}

func implementationHint(specPath, commandID, clientMethod string) string {
	if clientMethod == "" {
		return fmt.Sprintf("scaffolded CLI command %q from %s: wire implementation", commandID, specPath)
	}
	return fmt.Sprintf("scaffolded CLI command %q from %s: wire client method %s", commandID, specPath, clientMethod)
}

func goStringSlice(items []string) string {
	if len(items) == 0 {
		return "nil"
	}
	quoted := make([]string, 0, len(items))
	for _, item := range items {
		quoted = append(quoted, strconv.Quote(item))
	}
	return "[]string{" + strings.Join(quoted, ", ") + "}"
}

func snakeCase(s string) string {
	parts := identifierParts(s)
	if len(parts) == 0 {
		return "value"
	}
	for i := range parts {
		parts[i] = strings.ToLower(parts[i])
	}
	return strings.Join(parts, "_")
}

func packageNameFromTarget(targetFile string) string {
	dir := filepath.Base(filepath.Dir(targetFile))
	name := nonIdentifierCharsRE.ReplaceAllString(dir, "_")
	name = strings.Trim(name, "_")
	if name == "" {
		return "main"
	}
	if name[0] >= '0' && name[0] <= '9' {
		name = "pkg_" + name
	}
	return strings.ToLower(name)
}

func lowerCamel(s string) string {
	parts := identifierParts(s)
	if len(parts) == 0 {
		return "value"
	}
	var out strings.Builder
	out.WriteString(formatIdentifierPart(parts[0], true, false))
	for _, part := range parts[1:] {
		out.WriteString(formatIdentifierPart(part, false, false))
	}
	return out.String()
}

func pascal(s string) string {
	parts := identifierParts(s)
	if len(parts) == 0 {
		return "Generated"
	}
	var out strings.Builder
	for _, part := range parts {
		out.WriteString(formatIdentifierPart(part, false, true))
	}
	return out.String()
}

func identifierParts(s string) []string {
	rawParts := strings.Fields(nonIdentifierCharsRE.ReplaceAllString(s, " "))
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		parts = append(parts, splitCamelPart(part)...)
	}
	return parts
}

func splitCamelPart(part string) []string {
	var parts []string
	start := 0
	runes := []rune(part)
	for i := 1; i < len(runes); i++ {
		if isUpper(runes[i]) && (isLower(runes[i-1]) || isDigit(runes[i-1]) || (isUpper(runes[i-1]) && i+2 < len(runes) && isLower(runes[i+1]))) {
			parts = append(parts, string(runes[start:i]))
			start = i
		}
	}
	parts = append(parts, string(runes[start:]))
	return parts
}

func formatIdentifierPart(part string, first, exported bool) string {
	upper := strings.ToUpper(part)
	if commonInitialisms[upper] {
		if first && !exported {
			return strings.ToLower(upper)
		}
		return upper
	}
	lower := strings.ToLower(part)
	if first && !exported {
		return lower
	}
	return strings.ToUpper(lower[:1]) + lower[1:]
}

func safeGoIdentifier(name string) string {
	if name == "" {
		return "value"
	}
	if goKeywords[name] {
		return name + "Value"
	}
	if name[0] >= '0' && name[0] <= '9' {
		return "value" + name
	}
	return name
}

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func isLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

var commonInitialisms = map[string]bool{
	"API":  true,
	"CA":   true,
	"CPU":  true,
	"HTTP": true,
	"ID":   true,
	"JSON": true,
	"RAM":  true,
	"SHA":  true,
	"TLS":  true,
	"URL":  true,
	"UUID": true,
	"VM":   true,
}

var goKeywords = map[string]bool{
	"break":       true,
	"default":     true,
	"func":        true,
	"interface":   true,
	"select":      true,
	"case":        true,
	"defer":       true,
	"go":          true,
	"map":         true,
	"struct":      true,
	"chan":        true,
	"else":        true,
	"goto":        true,
	"package":     true,
	"switch":      true,
	"const":       true,
	"fallthrough": true,
	"if":          true,
	"range":       true,
	"type":        true,
	"continue":    true,
	"for":         true,
	"import":      true,
	"return":      true,
	"var":         true,
}
