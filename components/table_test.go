package components

import (
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout

	r, w, err := os.Pipe()

	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	os.Stdout = w

	fn()

	os.Stdout = original

	if err := w.Close(); err != nil {
		t.Fatalf("closing the pipe writer: %v", err)
	}

	data, err := io.ReadAll(r)

	if err != nil {
		t.Fatalf("reading the pipe: %v", err)
	}

	if err := r.Close(); err != nil {
		t.Fatalf("closing the pipe reader: %v", err)
	}

	return string(data)
}

func TestTableRendersTheTitleAndTheEmptyMessage(t *testing.T) {
	out := captureStdout(t, func() {
		table := Table{Title: "Commands", EmptyMsg: "No commands found"}
		table.Render()
	})

	if !strings.Contains(out, "Commands") {
		t.Fatalf("want the title, got %q", out)
	}
	if !strings.Contains(out, "No commands found") {
		t.Fatalf("want the empty message, got %q", out)
	}
}

func TestTableSkipsTheHeaderWhenThereAreNoRows(t *testing.T) {
	out := captureStdout(t, func() {
		table := Table{Header: []string{"Name"}, EmptyMsg: "empty"}
		table.Render()
	})

	if strings.Contains(out, "Name") {
		t.Fatalf("want no header, got %q", out)
	}
	if !strings.Contains(out, "empty") {
		t.Fatalf("want the empty message, got %q", out)
	}
}

func TestTableRendersTheHeaderAndTheRows(t *testing.T) {
	out := captureStdout(t, func() {
		table := Table{
			Header: []string{"Name", "Description"},
			Rows: [][]string{
				{"build", "compiles the project"},
				{"deploy", "ships it"},
			},
		}
		table.Render()
	})

	for _, want := range []string{"Name", "Description", "build", "compiles the project", "deploy", "ships it"} {
		if !strings.Contains(out, want) {
			t.Fatalf("want %q in the output, got %q", want, out)
		}
	}
}

func TestTableSeparatorUsesTheHeaderByteLength(t *testing.T) {
	out := captureStdout(t, func() {
		table := Table{
			Header: []string{"ação"},
			Rows:   [][]string{{"x"}},
		}
		table.Render()
	})

	if !strings.Contains(out, strings.Repeat("-", len("ação"))) {
		t.Fatalf("want a separator of %d dashes, got %q", len("ação"), out)
	}
}

func TestTableRendersWithoutAHeader(t *testing.T) {
	out := captureStdout(t, func() {
		table := Table{Rows: [][]string{{"name", "value"}}}
		table.Render()
	})

	if strings.Contains(out, "-") {
		t.Fatalf("want no separator, got %q", out)
	}
	if !strings.Contains(out, "name") || !strings.Contains(out, "value") {
		t.Fatalf("want the row, got %q", out)
	}
}
