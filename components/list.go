package components

import (
	"fmt"
	"os"
	"strings"
	"tartarus/themes"
	"text/tabwriter"
)

type List struct {
	Title    string
	EmptyMsg string
	Headers  []string
	Rows     [][]string
}

func (l *List) Print() {
	fmt.Printf("%s\n\n", themes.GetPrimaryText(l.Title, true))

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 3, ' ', 0)

	fmt.Fprintln(w, strings.Join(l.Headers, "\t"))
	fmt.Fprintln(w, makeTraceLine(l.Headers))

	if len(l.Rows) == 0 {
		fmt.Printf("%s\n\n", themes.GetTertiaryText(l.EmptyMsg, true))
		return
	}

	for _, row := range l.Rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	w.Flush()
}

func makeTraceLine(headers []string) string {
	tls := make([]string, len(headers))

	for i, h := range headers {
		tls[i] = strings.Repeat("-", len(h))
	}

	return strings.Join(tls, "\t") + "\t"
}
