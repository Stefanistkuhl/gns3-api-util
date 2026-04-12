package rgjson

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type MessageType string

const (
	TypeBegin   MessageType = "begin"
	TypeEnd     MessageType = "end"
	TypeMatch   MessageType = "match"
	TypeContext MessageType = "context"
	TypeSummary MessageType = "summary"
)

type TextOrBytes struct {
	Text  *string `json:"text,omitempty"`
	Bytes *string `json:"bytes,omitempty"`
}

func (t TextOrBytes) AsBytes() ([]byte, error) {
	if t.Text != nil {
		return []byte(*t.Text), nil
	}
	if t.Bytes != nil {
		return base64.StdEncoding.DecodeString(*t.Bytes)
	}
	return nil, nil
}

func (t TextOrBytes) String() string {
	if t.Text != nil {
		return *t.Text
	}
	if t.Bytes != nil {
		b, err := base64.StdEncoding.DecodeString(*t.Bytes)
		if err == nil {
			return string(b)
		}
	}
	return ""
}

type Path struct {
	TextOrBytes
}

type Lines struct {
	TextOrBytes
}

type MatchText struct {
	TextOrBytes
}

type Submatch struct {
	Match MatchText `json:"match"`
	Start int       `json:"start"`
	End   int       `json:"end"`
}

type Elapsed struct {
	Secs  int64  `json:"secs"`
	Nanos int64  `json:"nanos"`
	Human string `json:"human"`
}

type Stats struct {
	Elapsed           Elapsed `json:"elapsed"`
	Searches          int64   `json:"searches"`
	SearchesWithMatch int64   `json:"searches_with_match"`
	BytesSearched     int64   `json:"bytes_searched"`
	BytesPrinted      int64   `json:"bytes_printed"`
	MatchedLines      int64   `json:"matched_lines"`
	Matches           int64   `json:"matches"`
}

type BeginData struct {
	Path Path `json:"path"`
}

type MatchData struct {
	Path           Path       `json:"path"`
	Lines          Lines      `json:"lines"`
	LineNumber     *int64     `json:"line_number,omitempty"`
	AbsoluteOffset int64      `json:"absolute_offset"`
	Submatches     []Submatch `json:"submatches"`
}

type ContextData struct {
	Path           Path       `json:"path"`
	Lines          Lines      `json:"lines"`
	LineNumber     *int64     `json:"line_number,omitempty"`
	AbsoluteOffset int64      `json:"absolute_offset"`
	Submatches     []Submatch `json:"submatches,omitempty"`
}

type EndData struct {
	Path         Path   `json:"path"`
	BinaryOffset *int64 `json:"binary_offset"`
	Stats        Stats  `json:"stats"`
}

type SummaryData struct {
	ElapsedTotal Elapsed `json:"elapsed_total"`
	Stats        Stats   `json:"stats"`
}

type Envelope struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Event struct {
	Type    MessageType
	Begin   *BeginData
	Match   *MatchData
	Context *ContextData
	End     *EndData
	Summary *SummaryData
}

func (e Envelope) Decode() (Event, error) {
	ev := Event{Type: e.Type}

	switch e.Type {
	case TypeBegin:
		var v BeginData
		if err := json.Unmarshal(e.Data, &v); err != nil {
			return Event{}, fmt.Errorf("unmarshal begin: %w", err)
		}
		ev.Begin = &v
		return ev, nil

	case TypeMatch:
		var v MatchData
		if err := json.Unmarshal(e.Data, &v); err != nil {
			return Event{}, fmt.Errorf("unmarshal match: %w", err)
		}
		ev.Match = &v
		return ev, nil

	case TypeContext:
		var v ContextData
		if err := json.Unmarshal(e.Data, &v); err != nil {
			return Event{}, fmt.Errorf("unmarshal context: %w", err)
		}
		ev.Context = &v
		return ev, nil

	case TypeEnd:
		var v EndData
		if err := json.Unmarshal(e.Data, &v); err != nil {
			return Event{}, fmt.Errorf("unmarshal end: %w", err)
		}
		ev.End = &v
		return ev, nil

	case TypeSummary:
		var v SummaryData
		if err := json.Unmarshal(e.Data, &v); err != nil {
			return Event{}, fmt.Errorf("unmarshal summary: %w", err)
		}
		ev.Summary = &v
		return ev, nil

	default:
		return Event{}, fmt.Errorf("unknown ripgrep event type %q", e.Type)
	}
}

type Decoder struct {
	scanner *bufio.Scanner
}

func NewDecoder(r io.Reader) *Decoder {
	sc := bufio.NewScanner(r)

	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 10*1024*1024)

	return &Decoder{scanner: sc}
}

func (d *Decoder) Next() (Event, error) {
	for d.scanner.Scan() {
		line := d.scanner.Bytes()

		if len(line) == 0 {
			continue
		}

		var env Envelope
		if err := json.Unmarshal(line, &env); err != nil {
			return Event{}, fmt.Errorf("unmarshal envelope: %w", err)
		}

		ev, err := env.Decode()
		if err != nil {
			return Event{}, err
		}

		if ev.Type != "" {
			return ev, nil
		}
	}

	if err := d.scanner.Err(); err != nil {
		return Event{}, err
	}

	return Event{}, io.EOF
}

func DecodeAll(r io.Reader) ([]Event, error) {
	dec := NewDecoder(r)
	var out []Event

	for {
		ev, err := dec.Next()
		if errors.Is(err, io.EOF) {
			return out, nil
		}
		if err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
}

type LineLocation struct {
	Path       string
	LineNumber int64
}

type SectionInfo struct {
	Path  string
	Name  string
	Start int64
	End   int64
}

var sectionHeaderRE = regexp.MustCompile(`(?i)^// --- (.+?) ---\s*$`)

func FindAllSections(file string) ([]SectionInfo, error) {
	command := fmt.Sprintf(
		`rg -nUPi --json '^// --- (.+?) ---\s*$' %q`,
		file,
	)

	events, err := RunGrep(command)
	if err != nil {
		return nil, err
	}

	var sections []SectionInfo

	for _, event := range events {
		if event.Type != TypeMatch || event.Match == nil || event.Match.LineNumber == nil {
			continue
		}

		line := strings.TrimSpace(event.Match.Lines.String())
		m := sectionHeaderRE.FindStringSubmatch(line)
		if len(m) != 2 {
			continue
		}

		sections = append(sections, SectionInfo{
			Path:  file,
			Name:  strings.TrimSpace(m[1]),
			Start: *event.Match.LineNumber,
		})
	}

	if len(sections) == 0 {
		return nil, fmt.Errorf("no sections found in %s", file)
	}

	content, err := os.ReadFile(file) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("read file for line count: %w", err)
	}

	totalLines := int64(strings.Count(string(content), "\n"))
	if len(content) > 0 && !bytes.HasSuffix(content, []byte("\n")) {
		totalLines++
	}

	for i := range sections {
		if i+1 < len(sections) {
			sections[i].End = sections[i+1].Start - 1
		} else {
			sections[i].End = totalLines
		}
	}

	return sections, nil
}

func FindSection(name, file string) (SectionInfo, error) {
	sections, err := FindAllSections(file)
	if err != nil {
		return SectionInfo{}, err
	}

	for _, section := range sections {
		if section.Name == name {
			return section, nil
		}
	}

	return SectionInfo{}, fmt.Errorf("section %q not found in %s", name, file)
}

func FindStructDef(structName string) (LineLocation, error) {
	var dat LineLocation

	pattern := fmt.Sprintf("type %s struct {", structName)
	command := fmt.Sprintf(
		"rg -F --json -g '*.go' %q .",
		pattern,
	)

	events, err := RunGrep(command)
	if err != nil {
		return dat, err
	}

	for _, event := range events {
		if event.Type != TypeMatch || event.Match == nil || event.Match.LineNumber == nil {
			continue
		}

		path := event.Match.Path.String()
		if path == "" {
			continue
		}

		dat.Path = path
		dat.LineNumber = *event.Match.LineNumber
		return dat, nil
	}

	return dat, fmt.Errorf("struct definition not found: %s", structName)
}

func RunGrep(command string) ([]Event, error) {
	cmd := exec.CommandContext(context.Background(), "bash", "-c", command) // #nosec G204

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("rg failed: %w: %s", err, string(output))
	}

	return DecodeAll(bytes.NewReader(output))
}
