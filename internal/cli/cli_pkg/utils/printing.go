package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/globals"
	"github.com/charmbracelet/lipgloss"
	"github.com/pelletier/go-toml/v2"
	"github.com/tidwall/pretty"
	"gopkg.in/yaml.v3"
)

type ResourcePrinter interface {
	PrintObj(obj any, w io.Writer) error
}

type TableRecord interface {
	GetHeaders() []string
	GetRow() []string
}

func GetPrinter(format string) (ResourcePrinter, error) {
	switch format {
	case globals.OutputTable.String():
		return &LipglossTablePrinter{}, nil
	case "plain", "table-ascii":
		return &ASCIITablePrinter{}, nil
	case globals.OutputJSON.String():
		return &JSONColorPrinter{}, nil
	case globals.OutputJSONColorless.String():
		return &JSONPrinter{}, nil
	case globals.OutputCollapsed.String():
		return &JSONCompactPrinter{}, nil
	case globals.OutputYAML.String():
		return &YAMLPrinter{}, nil
	case globals.OutputTOML.String():
		return &TOMLPrinter{}, nil
	default:
		return &JSONColorPrinter{}, nil
	}
}

func PrintOutput(body []byte, cfg *config.GlobalOptions) {
	var data any
	_ = json.Unmarshal(body, &data)

	printer, err := GetPrinter(cfg.OutputFormat.String())
	if err != nil {
		printer = &JSONColorPrinter{}
	}

	if (cfg.OutputFormat == globals.OutputTable || cfg.OutputFormat == globals.OutputTableASCII) && !isTableCompatible(data) {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("⚠️  Command not yet updated for table output. Falling back to JSON:"))
		printer = &JSONColorPrinter{}
	}

	_ = printer.PrintObj(data, os.Stdout)
}

func isTableCompatible(obj any) bool {
	_, ok1 := obj.(TableRecord)
	_, ok2 := obj.([]TableRecord)
	return ok1 || ok2
}

type JSONColorPrinter struct{}

func (p *JSONColorPrinter) PrintObj(obj any, w io.Writer) error {
	raw, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	colored := pretty.Color(pretty.Pretty(raw), nil)
	_, err = fmt.Fprintln(w, string(colored))
	return err
}

type JSONPrinter struct{}

func (p *JSONPrinter) PrintObj(obj any, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(obj)
}

type JSONCompactPrinter struct{}

func (p *JSONCompactPrinter) PrintObj(obj any, w io.Writer) error {
	raw, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(raw))
	return err
}

type YAMLPrinter struct{}

func (p *YAMLPrinter) PrintObj(obj any, w io.Writer) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(obj)
}

type TOMLPrinter struct{}

func (p *TOMLPrinter) PrintObj(obj any, w io.Writer) error {
	enc := toml.NewEncoder(w)
	return enc.Encode(obj)
}

type ASCIITablePrinter struct{}

func (p *ASCIITablePrinter) PrintObj(obj any, w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 8, 3, ' ', 0)
	defer func() { _ = tw.Flush() }()

	switch items := obj.(type) {
	case []TableRecord:
		if len(items) == 0 {
			_, _ = fmt.Fprintln(w, "No resources found.")
			return nil
		}
		_, _ = fmt.Fprintln(tw, strings.Join(items[0].GetHeaders(), "\t"))
		for _, item := range items {
			_, _ = fmt.Fprintln(tw, strings.Join(item.GetRow(), "\t"))
		}
	case TableRecord:
		_, _ = fmt.Fprintln(tw, strings.Join(items.GetHeaders(), "\t"))
		_, _ = fmt.Fprintln(tw, strings.Join(items.GetRow(), "\t"))
	default:
		return fmt.Errorf("object does not implement TableRecord")
	}

	return nil
}

type LipglossTablePrinter struct{}

func (p *LipglossTablePrinter) PrintObj(obj any, w io.Writer) error {
	var originalHeaders []string
	var originalRows [][]string

	switch items := obj.(type) {
	case []TableRecord:
		if len(items) == 0 {
			_, _ = fmt.Fprintln(w, "No resources found.")
			return nil
		}
		originalHeaders = items[0].GetHeaders()
		for _, item := range items {
			originalRows = append(originalRows, item.GetRow())
		}
	case TableRecord:
		originalHeaders = items.GetHeaders()
		originalRows = append(originalRows, items.GetRow())
	default:
		return fmt.Errorf("object does not implement TableRecord")
	}

	headers := append([]string{"#"}, originalHeaders...)
	var rows [][]string
	for i, row := range originalRows {
		indexedRow := append([]string{fmt.Sprintf("%d", i)}, row...)
		rows = append(rows, indexedRow)
	}

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = lipgloss.Width(h)
	}
	for _, row := range rows {
		for i, col := range row {
			if width := lipgloss.Width(col); width > colWidths[i] {
				colWidths[i] = width
			}
		}
	}

	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Padding(0, 1)
	rowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Padding(0, 1)
	indexStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 1)
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	drawDivider := func(left, mid, right string) string {
		parts := make([]string, 0, len(colWidths))
		for _, w := range colWidths {
			// w + 2 accounts for the left and right padding inside the cell
			parts = append(parts, strings.Repeat("─", w+2))
		}
		return borderStyle.Render(left + strings.Join(parts, mid) + right)
	}

	drawRow := func(rowData []string, isHeader bool) string {
		parts := make([]string, 0, len(rowData))
		for i, colText := range rowData {
			style := rowStyle
			align := lipgloss.Left

			if isHeader {
				style = headerStyle
				align = lipgloss.Center
				if i == 0 {
					align = lipgloss.Right
				}
			} else if i == 0 {
				style = indexStyle
				align = lipgloss.Right
			}

			cell := style.Width(colWidths[i] + 2).Align(align).Render(colText)
			parts = append(parts, cell)
		}
		sep := borderStyle.Render("│")
		return sep + strings.Join(parts, sep) + sep
	}

	var out []string

	out = append(out, drawDivider("╭", "┬", "╮"), drawRow(headers, true), drawDivider("├", "┼", "┤"))

	for _, row := range rows {
		out = append(out, drawRow(row, false))
	}

	out = append(out, drawDivider("├", "┼", "┤"), drawRow(headers, true), drawDivider("╰", "┴", "╯"))

	_, err := fmt.Fprintln(w, strings.Join(out, "\n"))
	return err
}
