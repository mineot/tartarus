package themes

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

func TestTextHelpersRenderTheirValue(t *testing.T) {
	helpers := []struct {
		name string
		fn   func(string, bool) string
	}{
		{"primary", GetPrimaryText},
		{"secondary", GetSecondaryText},
		{"tertiary", GetTertiaryText},
		{"asset", GetAssetText},
	}

	for _, helper := range helpers {
		for _, bold := range []bool{false, true} {
			got := helper.fn("hello", bold)

			if !strings.Contains(got, "hello") {
				t.Fatalf("%s(bold=%v): want the value, got %q", helper.name, bold, got)
			}
		}
	}
}

func TestTextHelpersStyleTheValue(t *testing.T) {
	for _, got := range []string{
		GetPrimaryText("hello", false),
		GetSecondaryText("hello", false),
		GetTertiaryText("hello", false),
		GetAssetText("hello", false),
	} {
		if got == "hello" {
			t.Fatalf("want a styled value, got the raw text")
		}
	}
}

func TestRenderLogoWritesTheNameAndVersion(t *testing.T) {
	out := captureStdout(t, RenderLogo)

	if !strings.Contains(out, "Tartarus") {
		t.Fatalf("want the name, got %q", out)
	}
	if !strings.Contains(out, "2.0.0") {
		t.Fatalf("want the version, got %q", out)
	}
}
