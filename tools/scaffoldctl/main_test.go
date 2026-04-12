package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/0xveya/gns3util/tools/scaffoldctl/create"
	"github.com/0xveya/gns3util/tools/scaffoldctl/spec"
)

func TestParseFlagsNoPagerAndListModes(t *testing.T) {
	cfg, showHelp, err := parseFlags([]string{"-dry-run", "-no-pager", "-skip-existing", "-list-client", "-only", "get_backup,ListVMImages"})
	if err != nil {
		t.Fatal(err)
	}
	if showHelp {
		t.Fatal("showHelp should be false")
	}
	if !cfg.DryRun {
		t.Fatal("DryRun should be true")
	}
	if !cfg.NoPager {
		t.Fatal("NoPager should be true")
	}
	if !cfg.ListClient {
		t.Fatal("ListClient should be true")
	}
	if !cfg.SkipExisting {
		t.Fatal("SkipExisting should be true")
	}
	if want := []string{"get_backup", "ListVMImages"}; !reflect.DeepEqual(cfg.Only, want) {
		t.Fatalf("Only = %#v, want %#v", cfg.Only, want)
	}
}

func TestParseFlagsRejectsDryRunAndWrite(t *testing.T) {
	if _, _, err := parseFlags([]string{"-dry-run", "-write"}); err == nil {
		t.Fatal("expected -dry-run and -write to be rejected together")
	}
}

func TestParseFlagsSuggestsClosestFlagName(t *testing.T) {
	_, _, err := parseFlags([]string{"-chceck"})
	if err == nil {
		t.Fatal("expected unknown flag error")
	}
	if !strings.Contains(err.Error(), "did you mean -check?") {
		t.Fatalf("missing flag suggestion: %v", err)
	}
}

func TestSortedClientMethodsFiltersAndSorts(t *testing.T) {
	methods := []spec.ClientMethodSpec{
		{ID: "list_backups", Method: "ListBackups"},
		{ID: "delete_backup", Method: "DeleteBackup"},
		{ID: "get_backup", Method: "GetBackup"},
	}

	got := sortedClientMethods(methods, onlySet([]string{"get_backup", "deletebackup"}))
	if len(got) != 2 {
		t.Fatalf("len(sorted) = %d, want 2", len(got))
	}
	if got[0].Method != "DeleteBackup" || got[1].Method != "GetBackup" {
		t.Fatalf("sorted methods = %#v", got)
	}
}

func TestBuildClientFilesSkipsExistingMethods(t *testing.T) {
	target := filepath.Join(t.TempDir(), "client_v2.go")
	err := os.WriteFile(target, []byte(strings.Join([]string{
		"package api",
		"",
		"// --- Backups ---",
		"",
		"func (c *ClientV2) GetBackup(ctx context.Context, fileUUID string) error { return nil }",
		"",
	}, "\n")), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	specFile := &spec.SpecFile{
		Path: "backups.yaml",
		Package: spec.PackageMeta{
			ID: "backups",
		},
		Client: spec.ClientConfig{
			TargetFile:     target,
			DefaultSection: "backups",
		},
		ClientMethods: []spec.ClientMethodSpec{
			{
				ID:         "get_backup",
				Method:     "GetBackup",
				HTTPMethod: "GET",
				URLExpr:    `"/backups/existing"`,
				Response: spec.ResponseSpec{
					PointerType: "*models.BackupInfo",
					ValueType:   "models.BackupInfo",
					DecodeVar:   "resp",
					DecodeError: "failed to decode backup",
				},
			},
			{
				ID:         "list_backups",
				Method:     "ListBackups",
				HTTPMethod: "GET",
				URLExpr:    `"/backups"`,
				Response: spec.ResponseSpec{
					PointerType: "*models.ListBackupsResponse",
					ValueType:   "models.ListBackupsResponse",
					DecodeVar:   "resp",
					DecodeError: "failed to decode backups",
				},
			},
		},
	}

	builds, err := buildClientFiles([]*spec.SpecFile{specFile}, nil, "templates", true)
	if err != nil {
		t.Fatal(err)
	}

	generated, err := create.BuildFinalFile(builds[target])
	if err != nil {
		t.Fatal(err)
	}

	if strings.Count(generated, "GetBackup") != 1 {
		t.Fatalf("existing method should not be generated again:\n%s", generated)
	}
	if !strings.Contains(generated, "ListBackups") {
		t.Fatalf("missing method should still be generated:\n%s", generated)
	}
}

func TestFindExistingClientMethodsInText(t *testing.T) {
	text := `package api

// Scaffolded by tools/scaffoldctl; edit as needed.
func (c *ClientV2) ListBackups(ctx context.Context) error {
	return nil
}

func (x *Other) Ignore() {}

func (c *ClientV2) GetBackup(ctx context.Context) error {
	return nil
}
`

	got, err := findExistingClientMethodsInText("client_v2.go", text)
	if err != nil {
		t.Fatal(err)
	}
	if got["ListBackups"].Line != 3 {
		t.Fatalf("ListBackups line = %d, want 3", got["ListBackups"].Line)
	}
	if !got["ListBackups"].Managed {
		t.Fatalf("ListBackups should be managed: %#v", got["ListBackups"])
	}
	if got["GetBackup"].Line != 10 {
		t.Fatalf("GetBackup line = %d, want 10", got["GetBackup"].Line)
	}
	if got["GetBackup"].Managed {
		t.Fatalf("GetBackup should be human-owned: %#v", got["GetBackup"])
	}
	if _, ok := got["Ignore"]; ok {
		t.Fatalf("non-ClientV2 method should not be included: %#v", got)
	}
}

func TestBuildClientFilesReplacesManagedMethod(t *testing.T) {
	target := filepath.Join(t.TempDir(), "client_v2.go")
	err := os.WriteFile(target, []byte(strings.Join([]string{
		"package api",
		"",
		"// --- Backups ---",
		"",
		"// Scaffolded by tools/scaffoldctl; edit as needed.",
		"func (c *ClientV2) GetBackup(ctx context.Context, fileUUID string) (*models.BackupInfo, error) {",
		"\treturn nil, fmt.Errorf(\"old\")",
		"}",
		"",
	}, "\n")), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	specFile := &spec.SpecFile{
		Path:    "backups.yaml",
		Package: spec.PackageMeta{ID: "backups"},
		Client: spec.ClientConfig{
			TargetFile:     target,
			DefaultSection: "backups",
		},
		ClientMethods: []spec.ClientMethodSpec{
			{
				ID:         "get_backup",
				Method:     "GetBackup",
				HTTPMethod: "GET",
				Args: []spec.MethodArgSpec{
					{Name: "fileUUID", Type: "string"},
				},
				URLExpr: `fmt.Sprintf("/backups/%s", fileUUID)`,
				Response: spec.ResponseSpec{
					PointerType: "*models.BackupInfo",
					ValueType:   "models.BackupInfo",
					DecodeVar:   "resp",
					DecodeError: "failed to decode backup",
				},
			},
		},
	}

	builds, err := buildClientFiles([]*spec.SpecFile{specFile}, nil, "templates", false)
	if err != nil {
		t.Fatal(err)
	}

	got, err := create.BuildFinalFile(builds[target])
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(got, `fmt.Errorf("old")`) {
		t.Fatalf("managed method should have been replaced:\n%s", got)
	}
	if !strings.Contains(got, `fmt.Sprintf("/backups/%s", fileUUID)`) {
		t.Fatalf("new managed method body missing:\n%s", got)
	}
}

func TestWriteGeneratedNewFileRespectsScaffoldMarker(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backups.go")
	if err := os.WriteFile(path, []byte("// edited by human\npackage objstorecmd\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := writeGeneratedNewFile(path, "// Scaffolded by tools/scaffoldctl from spec\npackage objstorecmd\n", false)
	if err == nil {
		t.Fatal("expected overwrite of human-owned file to be rejected")
	}

	if err := os.WriteFile(path, []byte("// Scaffolded by tools/scaffoldctl from spec\npackage objstorecmd\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeGeneratedNewFile(path, "// Scaffolded by tools/scaffoldctl from spec\npackage objstorecmd\n// next\n", false); err != nil {
		t.Fatalf("managed scaffold file should update without force: %v", err)
	}
}
