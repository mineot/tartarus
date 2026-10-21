package components

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

type failingReader struct {
	err error
}

func (r failingReader) Read([]byte) (int, error) {
	return 0, r.err
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestPromptReturnsATrimmedValue(t *testing.T) {
	var out strings.Builder

	value, err := prompt(strings.NewReader("  hello  \n"), &out, PromptOpts{Label: "Name"})

	if err != nil {
		t.Fatalf("prompt: %v", err)
	}
	if value != "hello" {
		t.Fatalf("want %q, got %q", "hello", value)
	}
}

func TestPromptReturnsTheDefaultOnAnEmptyLine(t *testing.T) {
	var out strings.Builder

	value, err := prompt(strings.NewReader("\n"), &out, PromptOpts{Label: "Name", Default: "main"})

	if err != nil {
		t.Fatalf("prompt: %v", err)
	}
	if value != "main" {
		t.Fatalf("want %q, got %q", "main", value)
	}
}

func TestPromptShowsANonemptyDefaultInBrackets(t *testing.T) {
	var out strings.Builder

	if _, err := prompt(strings.NewReader("\n"), &out, PromptOpts{Label: "Name", Default: "main"}); err != nil {
		t.Fatalf("prompt: %v", err)
	}
	if !strings.Contains(out.String(), "Name [main]: ") {
		t.Fatalf("want the default in brackets, got %q", out.String())
	}
}

func TestPromptDoesNotShowAnEmptyDefaultInBrackets(t *testing.T) {
	var out strings.Builder

	if _, err := prompt(strings.NewReader("\n"), &out, PromptOpts{Label: "Name"}); err != nil {
		t.Fatalf("prompt: %v", err)
	}
	if out.String() != "Name: " {
		t.Fatalf("want %q, got %q", "Name: ", out.String())
	}
}

func TestPromptReturnsTheLastLineWithoutANewline(t *testing.T) {
	var out strings.Builder

	value, err := prompt(strings.NewReader("abc"), &out, PromptOpts{Label: "Name"})

	if err != nil {
		t.Fatalf("prompt: %v", err)
	}
	if value != "abc" {
		t.Fatalf("want %q, got %q", "abc", value)
	}
}

func TestPromptDropsCarriageReturns(t *testing.T) {
	var out strings.Builder

	value, err := prompt(strings.NewReader("ab\r\n"), &out, PromptOpts{Label: "Name"})

	if err != nil {
		t.Fatalf("prompt: %v", err)
	}
	if value != "ab" {
		t.Fatalf("want %q, got %q", "ab", value)
	}
}

func TestPromptRepromptsAfterARejectedValue(t *testing.T) {
	var out strings.Builder

	validate := func(value string) error {
		if value == "bad" {
			return errors.New("nope")
		}

		return nil
	}

	value, err := prompt(strings.NewReader("bad\nok\n"), &out, PromptOpts{Label: "Name", Validate: validate})

	if err != nil {
		t.Fatalf("prompt: %v", err)
	}
	if value != "ok" {
		t.Fatalf("want %q, got %q", "ok", value)
	}
	if !strings.Contains(out.String(), "\x1b[31mnope\x1b[0m") {
		t.Fatalf("want the red reason, got %q", out.String())
	}
	if !strings.Contains(out.String(), "Name [bad]: ") {
		t.Fatalf("want the rejected value kept as the new default, got %q", out.String())
	}
}

func TestPromptReturnsUnexpectedEOFOnAClosedInput(t *testing.T) {
	_, err := prompt(strings.NewReader(""), io.Discard, PromptOpts{Label: "Name"})

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("want %v, got %v", io.ErrUnexpectedEOF, err)
	}
}

func TestPromptReturnsAReadError(t *testing.T) {
	readErr := errors.New("boom")

	_, err := prompt(failingReader{err: readErr}, io.Discard, PromptOpts{Label: "Name"})

	if !errors.Is(err, readErr) {
		t.Fatalf("want %v, got %v", readErr, err)
	}
}

func TestPromptReturnsAWriteError(t *testing.T) {
	writeErr := errors.New("boom")

	_, err := prompt(strings.NewReader("x\n"), failingWriter{err: writeErr}, PromptOpts{Label: "Name"})

	if !errors.Is(err, writeErr) {
		t.Fatalf("want %v, got %v", writeErr, err)
	}
}

func TestConfirmDefaultsToNo(t *testing.T) {
	var out strings.Builder

	answer, err := confirm(strings.NewReader("\n"), &out, ConfirmOpts{Question: "Sure?"})

	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if answer {
		t.Fatalf("want false, got true")
	}
}

func TestConfirmDefaultsToYes(t *testing.T) {
	var out strings.Builder

	answer, err := confirm(strings.NewReader("\n"), &out, ConfirmOpts{Question: "Sure?", YesDefault: true})

	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if !answer {
		t.Fatalf("want true, got false")
	}
}

func TestConfirmAcceptsYesAndNo(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"y\n", true},
		{"Y\n", true},
		{"yes\n", true},
		{"YES\n", true},
		{"no\n", false},
		{"n\n", false},
		{"N\n", false},
		{"NO\n", false},
	}

	for _, tc := range cases {
		var out strings.Builder

		answer, err := confirm(strings.NewReader(tc.input), &out, ConfirmOpts{Question: "Sure?"})

		if err != nil {
			t.Fatalf("confirm(%q): %v", tc.input, err)
		}
		if answer != tc.want {
			t.Fatalf("confirm(%q): want %v, got %v", tc.input, tc.want, answer)
		}
	}
}

func TestConfirmRepromptsOnAnInvalidAnswer(t *testing.T) {
	var out strings.Builder

	answer, err := confirm(strings.NewReader("invalid\nmaybe\ny\n"), &out, ConfirmOpts{Question: "Sure?"})

	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if !answer {
		t.Fatalf("want true, got false")
	}
	if !strings.Contains(out.String(), "Please answer y or n") {
		t.Fatalf("want the red hint, got %q", out.String())
	}
}

func TestConfirmHintFollowsYesDefault(t *testing.T) {
	var no strings.Builder

	if _, err := confirm(strings.NewReader("\n"), &no, ConfirmOpts{Question: "Sure?"}); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if !strings.Contains(no.String(), "Sure? (y/N): ") {
		t.Fatalf("want the (y/N) hint, got %q", no.String())
	}

	var yes strings.Builder

	if _, err := confirm(strings.NewReader("\n"), &yes, ConfirmOpts{Question: "Sure?", YesDefault: true}); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if !strings.Contains(yes.String(), "Sure? (Y/n): ") {
		t.Fatalf("want the (Y/n) hint, got %q", yes.String())
	}
}

func TestConfirmReturnsUnexpectedEOFOnAClosedInput(t *testing.T) {
	_, err := confirm(strings.NewReader(""), io.Discard, ConfirmOpts{Question: "Sure?"})

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("want %v, got %v", io.ErrUnexpectedEOF, err)
	}
}

func TestConfirmReturnsAReadError(t *testing.T) {
	readErr := errors.New("boom")

	_, err := confirm(failingReader{err: readErr}, io.Discard, ConfirmOpts{Question: "Sure?"})

	if !errors.Is(err, readErr) {
		t.Fatalf("want %v, got %v", readErr, err)
	}
}

func TestConfirmReturnsAWriteError(t *testing.T) {
	writeErr := errors.New("boom")

	_, err := confirm(strings.NewReader("y\n"), failingWriter{err: writeErr}, ConfirmOpts{Question: "Sure?"})

	if !errors.Is(err, writeErr) {
		t.Fatalf("want %v, got %v", writeErr, err)
	}
}

func withStdin(t *testing.T, input string) {
	t.Helper()

	originalIn := os.Stdin
	originalErr := os.Stderr

	in, inW, err := os.Pipe()

	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	if _, err := inW.WriteString(input); err != nil {
		t.Fatalf("writing stdin: %v", err)
	}

	if err := inW.Close(); err != nil {
		t.Fatalf("closing the stdin writer: %v", err)
	}

	errR, errW, err := os.Pipe()

	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	os.Stdin = in
	os.Stderr = errW

	t.Cleanup(func() {
		os.Stdin = originalIn
		os.Stderr = originalErr
		in.Close()
		errR.Close()
		errW.Close()
	})
}

func TestPromptReadsFromStdin(t *testing.T) {
	withStdin(t, "value\n")

	value, err := Prompt(PromptOpts{Label: "Name"})

	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}
	if value != "value" {
		t.Fatalf("want %q, got %q", "value", value)
	}
}

func TestConfirmReadsFromStdin(t *testing.T) {
	withStdin(t, "y\n")

	answer, err := Confirm(ConfirmOpts{Question: "Sure?"})

	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if !answer {
		t.Fatalf("want true, got false")
	}
}

func TestPromptTitlePrintsTheTitle(t *testing.T) {
	out := captureStdout(t, func() {
		PromptTitle("Create a new command")
	})

	if !strings.Contains(out, "Create a new command") {
		t.Fatalf("want the title, got %q", out)
	}
}
