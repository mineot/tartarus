package components

import (
	"fmt"
	"os"
	"strings"
	"tartarus/themes"
	"text/tabwriter"

	"charm.land/lipgloss/v2"
)

type Table struct {
	Header   []string
	Rows     [][]string
	EmptyMsg string
	Title    string
	tw       *tabwriter.Writer
}

func (t *Table) Render() {
	fmt.Println(lipgloss.NewStyle().
		Foreground(lipgloss.Color(themes.PrimaryColor)).
		Render(t.Title))

	fmt.Println()

	if len(t.Rows) == 0 {
		fmt.Println(t.EmptyMsg)
		return
	}

	t.tw = tabwriter.NewWriter(os.Stdout, 1, 1, 2, ' ', 0)
	t.renderHeader()
	t.renderHeaderLines()
	t.renderRow()
	t.tw.Flush()
}

func (t *Table) renderHeader() {
	if len(t.Header) > 0 {
		fmt.Fprintln(t.tw, strings.Join(t.Header, "\t"))
	}
}

func (t *Table) renderHeaderLines() {
	if len(t.Header) > 0 {
		var lines []string

		for _, h := range t.Header {
			lines = append(lines, strings.Repeat("-", len(h)))
		}

		fmt.Fprintln(t.tw, strings.Join(lines, "\t"))
	}
}

func (t *Table) renderRow() {
	for _, r := range t.Rows {
		fmt.Fprintln(t.tw, strings.Join(r, "\t"))
	}
}
