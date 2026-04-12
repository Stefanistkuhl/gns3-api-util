package create

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/0xveya/gns3util/tools/scaffoldctl/spec"
)

func TestCreateClientMethodAddsBodyRequestArg(t *testing.T) {
	method := &spec.ClientMethodSpec{
		Method:     "InitBackupUpload",
		HTTPMethod: "POST",
		URLExpr:    `"/backups"`,
		Request: &spec.RequestSpec{
			Body:    true,
			ArgName: "req",
			Type:    "*models.InitBackupUploadRequest",
		},
		Response: spec.ResponseSpec{
			PointerType: "*models.InitUploadResponse",
			ValueType:   "models.InitUploadResponse",
			DecodeVar:   "resp",
			DecodeError: "failed to decode backup upload init response",
		},
	}

	rendered, err := CreateClientMethod(method, filepath.Join("..", "templates", "client_method.go.tmpl"))
	if err != nil {
		t.Fatal(err)
	}

	wantSignature := "func (c *ClientV2) InitBackupUpload(ctx context.Context, req *models.InitBackupUploadRequest) (*models.InitUploadResponse, error)"
	if !strings.Contains(rendered, wantSignature) {
		t.Fatalf("rendered method missing body request arg\nwant: %s\nrendered:\n%s", wantSignature, rendered)
	}
	if !strings.Contains(rendered, "json.Marshal(req)") {
		t.Fatalf("rendered method marshals the wrong request arg:\n%s", rendered)
	}
	if strings.Contains(rendered, `"+": %w"`) || strings.Contains(rendered, `+": %w"`) {
		t.Fatalf("rendered method still uses string concatenation for decode errors:\n%s", rendered)
	}
	if !strings.Contains(rendered, `fmt.Errorf("failed to decode backup upload init response: %w", err)`) {
		t.Fatalf("rendered method has the wrong decode error format:\n%s", rendered)
	}
}

func TestCreateCLIFileRendersGroupsAuthAndSafeFlagVars(t *testing.T) {
	specFile := &spec.SpecFile{
		Path: "backups.yaml",
		Package: spec.PackageMeta{
			ID: "backups",
		},
		CLI: spec.CLIConfig{
			TargetFile:     "internal/cli/cmds/ctlcmd/objstorecmd/backups.go",
			CommandFactory: "NewBackupsCmd",
			Use:            "backups",
			Aliases:        []string{"backup"},
			Short:          "Manage backups",
			Long:           "List, inspect, upload, and update backup metadata.",
			Example:        "  gns3util ctl object-store backups list",
			AuthMode:       "cluster-only",
			Groups: []spec.CLIGroupSpec{
				{ID: "query", Title: "Query commands:"},
				{ID: "action", Title: "Action commands:"},
			},
			Rows: []spec.CLIRowSpec{
				{
					Type: "backupRow",
					Fields: []spec.CLIRowFieldSpec{
						{Name: "FileUUID", Header: "FILE UUID"},
						{Name: "Compressed", Type: "bool", JSON: "is_compressed", Header: "COMPRESSED"},
						{Name: "CreatedAt", Header: "CREATED AT", Value: "fmt.Sprintf(\"created:%s\", r.CreatedAt)"},
					},
				},
			},
		},
	}
	commands := []spec.CLICommandSpec{
		{
			ID:           "upload_backup",
			Factory:      "newUploadBackupCmd",
			Use:          "upload <file-path>",
			Short:        "Upload a backup",
			Example:      "  gns3util ctl object-store backups upload backup.tar --backup-type full",
			GroupID:      "action",
			CobraArgs:    "cobra.ExactArgs(1)",
			ClientMethod: "InitBackupUpload",
			Flags: []spec.CLIFlagSpec{
				{
					Name:       "type",
					Shorthand:  "t",
					Type:       "string",
					Default:    "application/octet-stream",
					Usage:      "Explicit Content-Type for the backup file",
					ModelField: "ContentType",
				},
				{
					Name:       "backup-type",
					Type:       "string",
					Usage:      "Backup type",
					Required:   true,
					ModelField: "BackupType",
				},
			},
		},
	}

	rendered, err := CreateCLIFile(specFile, commands, filepath.Join("..", "templates", "cli_file.go.tmpl"))
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"package objstorecmd",
		"func NewBackupsCmd() *cobra.Command",
		`Aliases: []string{"backup"}`,
		`Long:    "List, inspect, upload, and update backup metadata."`,
		`Example: "  gns3util ctl object-store backups list"`,
		`&cobra.Group{ID: "query", Title: "Query commands:"}`,
		"type backupRow struct",
		"FileUUID string `json:\"file_uuid\"`",
		"Compressed bool `json:\"is_compressed\"`",
		"CreatedAt string `json:\"created_at\"`",
		`"FILE UUID"`,
		`"COMPRESSED"`,
		`"CREATED AT"`,
		`fmt.Sprintf("created:%s", r.CreatedAt)`,
		`Example: "  gns3util ctl object-store backups upload backup.tar --backup-type full"`,
		`"auth-mode": "cluster-only"`,
		"contentType string",
		`cmd.Flags().StringVarP(&contentType, "type", "t", "application/octet-stream", "Explicit Content-Type for the backup file")`,
		`cmd.MarkFlagRequired("backup-type")`,
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered CLI file missing %q:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, "\ttype string") {
		t.Fatalf("reserved Go keyword was used as a flag variable:\n%s", rendered)
	}
}

func TestBuildCLIParentFileAppendsMissingRegistrationBeforeReturn(t *testing.T) {
	target := writeTempTarget(t, strings.Join([]string{
		"package objstorecmd",
		"",
		"import \"github.com/spf13/cobra\"",
		"",
		"func NewObjCmd() *cobra.Command {",
		"\tcmd := &cobra.Command{Use: \"object-store\"}",
		"",
		"\tuploadCmd := NewUploadCmd()",
		"\tuploadCmd.GroupID = \"object\"",
		"\tcmd.AddCommand(uploadCmd)",
		"",
		"\treturn cmd",
		"}",
		"",
	}, "\n"))

	got, err := BuildCLIParentFile(target, []CLIParentRegistration{
		{ParentFactory: "NewObjCmd", CommandFactory: "NewBackupsCmd", GroupID: "object"},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"backupsCmd := NewBackupsCmd()",
		`backupsCmd.GroupID = "object"`,
		"cmd.AddCommand(\n\t\tbackupsCmd,\n\t)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("parent file missing %q:\n%s", want, got)
		}
	}
	assertOrder(t, got, "backupsCmd := NewBackupsCmd()", "\treturn cmd")
}

func TestBuildCLIParentFileSkipsExistingRegistration(t *testing.T) {
	target := writeTempTarget(t, strings.Join([]string{
		"package objstorecmd",
		"",
		"import \"github.com/spf13/cobra\"",
		"",
		"func NewObjCmd() *cobra.Command {",
		"\tcmd := &cobra.Command{Use: \"object-store\"}",
		"\tbackupsCmd := NewBackupsCmd()",
		"\tcmd.AddCommand(backupsCmd)",
		"\treturn cmd",
		"}",
		"",
	}, "\n"))

	got, err := BuildCLIParentFile(target, []CLIParentRegistration{
		{ParentFactory: "NewObjCmd", CommandFactory: "NewBackupsCmd", GroupID: "object"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if strings.Count(got, "NewBackupsCmd()") != 1 {
		t.Fatalf("existing parent registration should not be duplicated:\n%s", got)
	}
}

func TestBuildFinalFileAppendsToExistingSection(t *testing.T) {
	target := writeTempTarget(t, strings.Join([]string{
		"package api",
		"",
		"// --- Files ---",
		"",
		"func existingFileMethod() {}",
		"",
		"// --- Transfers ---",
		"",
		"func existingTransferMethod() {}",
		"",
	}, "\n"))

	build := NewFileBuild(target)
	build.Append("files", "func generatedFileMethod() {}")

	got, err := BuildFinalFile(build)
	if err != nil {
		t.Fatal(err)
	}

	assertOrder(t, got, "func existingFileMethod() {}", "func generatedFileMethod() {}")
	assertOrder(t, got, "func generatedFileMethod() {}", "// --- Transfers ---")
	if !strings.Contains(got, "func existingTransferMethod() {}") {
		t.Fatalf("unmatched section content was not preserved:\n%s", got)
	}
}

func TestBuildFinalFileAppendsToLastDuplicateSection(t *testing.T) {
	target := writeTempTarget(t, strings.Join([]string{
		"package api",
		"",
		"// --- Transfers ---",
		"",
		"func firstTransferMethod() {}",
		"",
		"// --- Transfers ---",
		"",
		"func secondTransferMethod() {}",
		"",
	}, "\n"))

	build := NewFileBuild(target)
	build.Append("transfers", "func generatedTransferMethod() {}")

	got, err := BuildFinalFile(build)
	if err != nil {
		t.Fatal(err)
	}

	assertOrder(t, got, "func firstTransferMethod() {}", "// --- Transfers ---\n\nfunc secondTransferMethod() {}")
	assertOrder(t, got, "func secondTransferMethod() {}", "func generatedTransferMethod() {}")
}

func TestBuildFinalFileAddsMissingSection(t *testing.T) {
	target := writeTempTarget(t, strings.Join([]string{
		"package api",
		"",
		"// --- Files ---",
		"",
		"func existingFileMethod() {}",
		"",
	}, "\n"))

	build := NewFileBuild(target)
	build.Append("file_permissions", "func generatedPermissionMethod() {}")

	got, err := BuildFinalFile(build)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "// --- File permissions ---") {
		t.Fatalf("missing section was not added with the configured display name:\n%s", got)
	}
	assertOrder(t, got, "func existingFileMethod() {}", "// --- File permissions ---")
	assertOrder(t, got, "// --- File permissions ---", "func generatedPermissionMethod() {}")
}

func TestBuildFinalFileSortsGeneratedMethods(t *testing.T) {
	target := writeTempTarget(t, strings.Join([]string{
		"package api",
		"",
		"// --- Backups ---",
		"",
	}, "\n"))

	build := NewFileBuild(target)
	build.AppendSorted("backups", "ListBackups", "func (c *ClientV2) ListBackups() {}")
	build.AppendSorted("backups", "GetBackup", "func (c *ClientV2) GetBackup() {}")
	build.AppendSorted("backups", "DeleteBackup", "func (c *ClientV2) DeleteBackup() {}")

	got, err := BuildFinalFile(build)
	if err != nil {
		t.Fatal(err)
	}

	assertOrder(t, got, "func (c *ClientV2) DeleteBackup() {}", "func (c *ClientV2) GetBackup() {}")
	assertOrder(t, got, "func (c *ClientV2) GetBackup() {}", "func (c *ClientV2) ListBackups() {}")
}

func TestBuildFinalFileReplacesManagedMethodOnly(t *testing.T) {
	target := writeTempTarget(t, strings.Join([]string{
		"package api",
		"",
		"// --- Backups ---",
		"",
		"// Scaffolded by tools/scaffoldctl; edit as needed.",
		"func (c *ClientV2) GetBackup() {",
		"\tprintln(\"old\")",
		"}",
		"",
		"func (c *ClientV2) DeleteBackup() {",
		"\tprintln(\"human\")",
		"}",
		"",
	}, "\n"))

	build := NewFileBuild(target)
	build.Replace("GetBackup", "func (c *ClientV2) GetBackup() {\n\tprintln(\"new\")\n}")

	got, err := BuildFinalFile(build)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, `println("new")`) {
		t.Fatalf("managed method was not replaced:\n%s", got)
	}
	if !strings.Contains(got, `println("human")`) {
		t.Fatalf("human-owned method should have been left alone:\n%s", got)
	}
}

func writeTempTarget(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "client_v2.go")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertOrder(t *testing.T, haystack, before, after string) {
	t.Helper()

	beforeIndex := strings.Index(haystack, before)
	if beforeIndex < 0 {
		t.Fatalf("missing %q in:\n%s", before, haystack)
	}
	afterIndex := strings.Index(haystack, after)
	if afterIndex < 0 {
		t.Fatalf("missing %q in:\n%s", after, haystack)
	}
	if beforeIndex >= afterIndex {
		t.Fatalf("%q should appear before %q in:\n%s", before, after, haystack)
	}
}
