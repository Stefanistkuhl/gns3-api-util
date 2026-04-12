package create

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/0xveya/gns3util/tools/scaffoldctl/spec"
)

type BuildSection struct {
	Name    string
	Entries []BuildEntry
}

type BuildEntry struct {
	SortKey string
	Lines   []string
}

type FileBuild struct {
	TargetFile   string
	Sections     map[string]*BuildSection
	Replacements map[string]string
}

type foundSection struct {
	Name  string
	Start int
	End   int
}

var sectionHeaderRE = regexp.MustCompile(`(?i)^// --- (.+?) ---\s*$`)

var sectionDisplayNames = map[string]string{
	"transport":          "Transport",
	"auth":               "Auth",
	"cluster":            "Cluster",
	"users":              "Users",
	"roles":              "Roles",
	"jobs":               "Jobs",
	"buckets":            "Buckets",
	"files":              "Files",
	"backups":            "Backups",
	"vm images":          "VM images",
	"bucket permissions": "Bucket permissions",
	"file permissions":   "File permissions",
	"public tokens":      "Public tokens",
	"transfers":          "Transfers",
}

func NewFileBuild(targetFile string) *FileBuild {
	return &FileBuild{
		TargetFile:   targetFile,
		Sections:     map[string]*BuildSection{},
		Replacements: map[string]string{},
	}
}

func (b *FileBuild) Append(sectionName, rendered string) {
	b.AppendSorted(sectionName, "", rendered)
}

func (b *FileBuild) AppendSorted(sectionName, sortKey, rendered string) {
	sectionName = normalizeSectionName(sectionName)
	lines := splitLines(rendered)

	sec, ok := b.Sections[sectionName]
	if !ok {
		sec = &BuildSection{Name: sectionName}
		b.Sections[sectionName] = sec
	}

	sec.Entries = append(sec.Entries, BuildEntry{
		SortKey: strings.ToLower(strings.TrimSpace(sortKey)),
		Lines:   lines,
	})
}

func (b *FileBuild) Replace(methodName, rendered string) {
	if b == nil {
		return
	}
	if b.Replacements == nil {
		b.Replacements = map[string]string{}
	}
	b.Replacements[strings.TrimSpace(methodName)] = rendered
}

func CreateClientMethod(methodSpec *spec.ClientMethodSpec, templatePath string) (string, error) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, methodSpec); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func normalizeSectionName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.Join(strings.Fields(s), " ")
	return strings.ToLower(s)
}

func displaySectionName(s string) string {
	s = normalizeSectionName(s)
	if displayName, ok := sectionDisplayNames[s]; ok {
		return displayName
	}
	if s == "" {
		return ""
	}

	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func BuildFinalFile(build *FileBuild) (string, error) {
	var lines []string

	src, err := os.ReadFile(build.TargetFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("read target file %s: %w", build.TargetFile, err)
		}
		lines = []string{}
	} else {
		text := strings.ReplaceAll(string(src), "\r\n", "\n")
		text, err = applyManagedMethodReplacements(build.TargetFile, text, build.Replacements)
		if err != nil {
			return "", err
		}
		lines = strings.Split(text, "\n")
		if len(lines) == 1 && lines[0] == "" {
			lines = []string{}
		}
	}

	sections := findSections(lines)

	generatedByName := make(map[string]*BuildSection, len(build.Sections))
	for name, sec := range build.Sections {
		key := normalizeSectionName(name)
		if sec == nil {
			continue
		}
		sec.Name = key
		generatedByName[key] = sec
	}

	sort.Slice(sections, func(i, j int) bool {
		return sections[i].Start < sections[j].Start
	})

	lastSectionIndexByName := make(map[string]int, len(sections))
	for i, sec := range sections {
		lastSectionIndexByName[normalizeSectionName(sec.Name)] = i
	}

	var out []string
	cursor := 1

	for i, sec := range sections {
		startIdx := sec.Start - 1
		endIdx := sec.End - 1

		if startIdx >= cursor-1 && startIdx <= len(lines) {
			out = append(out, lines[cursor-1:startIdx]...)
		}

		if startIdx >= 0 && startIdx < len(lines) {
			out = append(out, lines[startIdx])
		}

		key := normalizeSectionName(sec.Name)
		if startIdx+1 <= endIdx && endIdx < len(lines) {
			out = append(out, lines[startIdx+1:endIdx+1]...)
		}

		if gen, ok := generatedByName[key]; ok && lastSectionIndexByName[key] == i {
			out = appendGeneratedLines(out, gen.Lines())
			delete(generatedByName, key)
		}

		cursor = endIdx + 2
	}

	if cursor-1 < len(lines) {
		out = append(out, lines[cursor-1:]...)
	}

	if len(generatedByName) > 0 {
		if len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
			out = append(out, "")
		}

		keys := make([]string, 0, len(generatedByName))
		for name := range generatedByName {
			keys = append(keys, name)
		}
		sort.Strings(keys)

		for i, key := range keys {
			sec := generatedByName[key]
			out = append(out, fmt.Sprintf("// --- %s ---", displaySectionName(key)))
			out = appendGeneratedLines(out, sec.Lines())
			if i < len(keys)-1 {
				out = append(out, "")
			}
		}
	}

	return strings.Join(out, "\n"), nil
}

func (s *BuildSection) Lines() []string {
	if s == nil || len(s.Entries) == 0 {
		return nil
	}

	entries := append([]BuildEntry(nil), s.Entries...)
	sort.SliceStable(entries, func(i, j int) bool {
		left := entries[i].SortKey
		right := entries[j].SortKey
		if left == "" {
			left = strings.Join(entries[i].Lines, "\n")
		}
		if right == "" {
			right = strings.Join(entries[j].Lines, "\n")
		}
		return left < right
	})

	var lines []string
	for i, entry := range entries {
		if i > 0 && len(entry.Lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, entry.Lines...)
	}
	return lines
}

func findSections(lines []string) []foundSection {
	totalLines := len(lines)
	if totalLines > 0 && lines[totalLines-1] == "" {
		totalLines--
	}

	var sections []foundSection
	for i := 0; i < totalLines; i++ {
		matches := sectionHeaderRE.FindStringSubmatch(strings.TrimSpace(lines[i]))
		if len(matches) != 2 {
			continue
		}

		sections = append(sections, foundSection{
			Name:  strings.TrimSpace(matches[1]),
			Start: i + 1,
		})
	}

	for i := range sections {
		if i+1 < len(sections) {
			sections[i].End = sections[i+1].Start - 1
			continue
		}
		sections[i].End = totalLines
	}

	return sections
}

func appendGeneratedLines(out, lines []string) []string {
	if len(lines) == 0 {
		return out
	}
	if len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
		out = append(out, "")
	}
	return append(out, lines...)
}
