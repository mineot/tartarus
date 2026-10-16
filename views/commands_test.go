package views

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tartarus/helpers"
	"tartarus/repositories"
	"tartarus/store"
)

func newTestRepos(t *testing.T) *repositories.Repos {
	t.Helper()

	if err := helpers.SetDevStorePath(filepath.Join(t.TempDir(), "test.db")); err != nil {
		t.Fatalf("SetDevStorePath: %v", err)
	}

	s, err := store.New(context.Background())

	if err != nil {
		t.Fatalf("store.New: %v", err)
	}

	t.Cleanup(func() { s.Close() })

	if err := s.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	return repositories.New(s)
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

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	original := os.Stdout

	r, w, err := os.Pipe()

	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	os.Stdout = w

	fnErr := fn()

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

	return string(data), fnErr
}

func runCommands(t *testing.T, r *repositories.Repos, input string, args ...string) (string, error) {
	t.Helper()

	withStdin(t, input)

	return captureStdout(t, func() error {
		cmd := NewCommands(r)
		cmd.SetArgs(args)

		return cmd.Execute()
	})
}

func insertTestCommand(t *testing.T, r *repositories.Repos, name string, description string) repositories.Command {
	t.Helper()

	command := repositories.Command{Name: name, Description: description}

	if err := r.InsertCommand(&command); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	return command
}

func TestNewRegistersTheCommandsSubcommand(t *testing.T) {
	r := newTestRepos(t)

	root := New(r)

	if root.Version != "2.0.0" {
		t.Fatalf("want version %q, got %q", "2.0.0", root.Version)
	}

	for _, c := range root.Commands() {
		if c.Name() == "cmds" {
			return
		}
	}

	t.Fatalf("want a cmds subcommand")
}

func TestRootCommandShowsHelp(t *testing.T) {
	r := newTestRepos(t)

	out, err := captureStdout(t, func() error {
		root := New(r)
		root.SetArgs([]string{})

		return root.Execute()
	})

	if err != nil {
		t.Fatalf("tartarus: %v", err)
	}
	if !strings.Contains(out, "Usage") {
		t.Fatalf("want the help, got %q", out)
	}
}

func TestCommandsRootShowsHelp(t *testing.T) {
	r := newTestRepos(t)

	out, err := runCommands(t, r, "")

	if err != nil {
		t.Fatalf("cmds: %v", err)
	}
	if !strings.Contains(out, "Usage") {
		t.Fatalf("want the help, got %q", out)
	}
}

func TestViewCommandRejectsTheWrongArgumentCount(t *testing.T) {
	r := newTestRepos(t)

	for _, args := range [][]string{{"v"}, {"v", "a", "b"}} {
		if _, err := runCommands(t, r, "", args...); err == nil {
			t.Fatalf("want an error for args %v", args)
		}
	}
}

func TestEditCommandRejectsTheWrongArgumentCount(t *testing.T) {
	r := newTestRepos(t)

	if _, err := runCommands(t, r, "", "e"); err == nil {
		t.Fatalf("want an error for a missing name")
	}
}

func TestDeleteCommandRejectsTheWrongArgumentCount(t *testing.T) {
	r := newTestRepos(t)

	if _, err := runCommands(t, r, "", "d"); err == nil {
		t.Fatalf("want an error for a missing name")
	}
}

func TestListCommandsRendersTheCommands(t *testing.T) {
	r := newTestRepos(t)
	insertTestCommand(t, r, "build", "compiles")

	out, err := runCommands(t, r, "", "ls")

	if err != nil {
		t.Fatalf("ls: %v", err)
	}
	if !strings.Contains(out, "build") || !strings.Contains(out, "compiles") {
		t.Fatalf("want the command in the table, got %q", out)
	}
}

func TestListCommandsRejectsABlankFilter(t *testing.T) {
	r := newTestRepos(t)

	_, err := runCommands(t, r, "", "ls", "--filter", "   ")

	if err == nil {
		t.Fatalf("want an error for a blank filter")
	}
}

func TestListCommandsFiltersByName(t *testing.T) {
	r := newTestRepos(t)
	insertTestCommand(t, r, "build", "compiles")
	insertTestCommand(t, r, "deploy", "ships")

	out, err := runCommands(t, r, "", "ls", "--filter", "depl")

	if err != nil {
		t.Fatalf("ls: %v", err)
	}
	if !strings.Contains(out, "deploy") {
		t.Fatalf("want the matching command, got %q", out)
	}
	if strings.Contains(out, "build") {
		t.Fatalf("want the other command filtered out, got %q", out)
	}
}

func TestViewCommandRendersTheCommandAndItsScripts(t *testing.T) {
	r := newTestRepos(t)
	command := insertTestCommand(t, r, "build", "compiles")

	script := repositories.CommandScript{Name: "gobuild", Script: "go build ./...", Description: "compiles"}

	if err := r.AppendCommandScript(command.ID, &script); err != nil {
		t.Fatalf("AppendCommandScript: %v", err)
	}

	out, err := runCommands(t, r, "", "v", "build")

	if err != nil {
		t.Fatalf("v: %v", err)
	}
	for _, want := range []string{"build", "compiles", "gobuild", "go build ./..."} {
		if !strings.Contains(out, want) {
			t.Fatalf("want %q in the output, got %q", want, out)
		}
	}
}

func TestCreateCommandInsertsACommand(t *testing.T) {
	r := newTestRepos(t)

	out, err := runCommands(t, r, "deploy\nship it\n", "c")

	if err != nil {
		t.Fatalf("c: %v", err)
	}

	got, err := r.GetCommand("deploy")

	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.ID == 0 {
		t.Fatalf("want the command inserted, got %+v", got)
	}
	if got.Description != "ship it" {
		t.Fatalf("want description %q, got %q", "ship it", got.Description)
	}
	if !strings.Contains(out, "deploy") {
		t.Fatalf("want the name in the output, got %q", out)
	}
}

func TestEditCommandUpdatesACommand(t *testing.T) {
	r := newTestRepos(t)
	insertTestCommand(t, r, "build", "old")

	if _, err := runCommands(t, r, "builder\nnew description\n", "e", "build"); err != nil {
		t.Fatalf("e: %v", err)
	}

	got, err := r.GetCommand("builder")

	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.ID == 0 {
		t.Fatalf("want the command renamed, got %+v", got)
	}
	if got.Description != "new description" {
		t.Fatalf("want description %q, got %q", "new description", got.Description)
	}
}

func TestDeleteCommandRemovesACommandOnConfirmation(t *testing.T) {
	r := newTestRepos(t)
	insertTestCommand(t, r, "build", "compiles")

	if _, err := runCommands(t, r, "y\n", "d", "build"); err != nil {
		t.Fatalf("d: %v", err)
	}

	got, err := r.GetCommand("build")

	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.ID != 0 {
		t.Fatalf("want the command deleted, got %+v", got)
	}
}

func TestDeleteCommandKeepsACommandWhenDeclined(t *testing.T) {
	r := newTestRepos(t)
	insertTestCommand(t, r, "build", "compiles")

	out, err := runCommands(t, r, "no\n", "d", "build")

	if err != nil {
		t.Fatalf("d: %v", err)
	}

	got, err := r.GetCommand("build")

	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.ID == 0 {
		t.Fatalf("want the command kept")
	}
	if !strings.Contains(out, "not deleted") {
		t.Fatalf("want the decline message, got %q", out)
	}
}

func TestDeleteCommandReturnsAConfirmationError(t *testing.T) {
	r := newTestRepos(t)
	insertTestCommand(t, r, "build", "compiles")

	if _, err := runCommands(t, r, "maybe\n", "d", "build"); err == nil {
		t.Fatalf("want the confirmation error to propagate")
	}
}

func TestCreateCommandRepromptsOnAnEmptyDescription(t *testing.T) {
	r := newTestRepos(t)

	if _, err := runCommands(t, r, "build\n\ndescription\n", "c"); err != nil {
		t.Fatalf("c: %v", err)
	}

	got, err := r.GetCommand("build")

	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.ID == 0 {
		t.Fatalf("want the command inserted, got %+v", got)
	}
	if got.Description != "description" {
		t.Fatalf("want description %q, got %q", "description", got.Description)
	}
}

func TestCreateCommandRejectsADuplicateName(t *testing.T) {
	r := newTestRepos(t)
	insertTestCommand(t, r, "build", "compiles")

	if _, err := runCommands(t, r, "build\nmore\n", "c"); err == nil {
		t.Fatalf("want the duplicate to fail")
	}
}

func TestCreateCommandReturnsAPromptError(t *testing.T) {
	r := newTestRepos(t)

	if _, err := runCommands(t, r, "", "c"); err == nil {
		t.Fatalf("want the name prompt error to propagate")
	}
}

func TestCreateCommandReturnsADescriptionPromptError(t *testing.T) {
	r := newTestRepos(t)

	if _, err := runCommands(t, r, "build\n", "c"); err == nil {
		t.Fatalf("want the description prompt error to propagate")
	}
}

func TestEditCommandReturnsAPromptError(t *testing.T) {
	r := newTestRepos(t)
	insertTestCommand(t, r, "build", "compiles")

	if _, err := runCommands(t, r, "", "e", "build"); err == nil {
		t.Fatalf("want the prompt error to propagate")
	}
}

func TestEditCommandRejectsADuplicateName(t *testing.T) {
	r := newTestRepos(t)
	insertTestCommand(t, r, "build", "compiles")
	insertTestCommand(t, r, "deploy", "ships")

	if _, err := runCommands(t, r, "deploy\nnew description\n", "e", "build"); err == nil {
		t.Fatalf("want the duplicate to fail")
	}
}

func TestEditCommandRepromptsOnAnEmptyDescription(t *testing.T) {
	r := newTestRepos(t)
	insertTestCommand(t, r, "build", "")

	if _, err := runCommands(t, r, "build\n\ndescription\n", "e", "build"); err != nil {
		t.Fatalf("e: %v", err)
	}

	got, err := r.GetCommand("build")

	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.Description != "description" {
		t.Fatalf("want description %q, got %q", "description", got.Description)
	}
}

func TestEditCommandReturnsADescriptionPromptError(t *testing.T) {
	r := newTestRepos(t)
	insertTestCommand(t, r, "build", "compiles")

	if _, err := runCommands(t, r, "builder\n", "e", "build"); err == nil {
		t.Fatalf("want the description prompt error to propagate")
	}
}

func TestEditCommandPropagatesAClosedStore(t *testing.T) {
	r := newTestRepos(t)

	if err := r.Store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := runCommands(t, r, "builder\nnew description\n", "e", "build"); err == nil {
		t.Fatalf("want the closed store error to propagate")
	}
}

func TestListCommandsPropagatesAClosedStore(t *testing.T) {
	r := newTestRepos(t)

	if err := r.Store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := runCommands(t, r, "", "ls"); err == nil {
		t.Fatalf("want the closed store error to propagate")
	}
}

func TestViewCommandPropagatesAClosedStore(t *testing.T) {
	r := newTestRepos(t)

	if err := r.Store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := runCommands(t, r, "", "v", "build"); err == nil {
		t.Fatalf("want the closed store error to propagate")
	}
}

func TestDeleteCommandPropagatesAClosedStore(t *testing.T) {
	r := newTestRepos(t)

	if err := r.Store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := runCommands(t, r, "y\n", "d", "build"); err == nil {
		t.Fatalf("want the closed store error to propagate")
	}
}
