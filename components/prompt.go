package components

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const promptRed = "\x1b[31m"
const promptReset = "\x1b[0m"

type PromptOpts struct {
	Label    string
	Default  string
	Validate func(string) error
}

type ConfirmOpts struct {
	Question string
	Default  bool
}

func Prompt(opts PromptOpts) (string, error) {
	return prompt(os.Stdin, os.Stderr, opts)
}

func Confirm(opts ConfirmOpts) (bool, error) {
	return confirm(os.Stdin, os.Stderr, opts)
}

func prompt(r io.Reader, w io.Writer, opts PromptOpts) (string, error) {
	value := opts.Default

	for {
		label := opts.Label

		if value != "" {
			label = fmt.Sprintf("%s [%s]", opts.Label, value)
		}

		if _, err := fmt.Fprintf(w, "%s: ", label); err != nil {
			return "", fmt.Errorf("components: prompt: %w", err)
		}

		line, err := readLine(r)

		if err != nil {
			return "", fmt.Errorf("components: prompt: %w", err)
		}

		line = strings.TrimSpace(line)

		if line == "" {
			line = value
		}

		if opts.Validate != nil {
			if err := opts.Validate(line); err != nil {
				if _, werr := fmt.Fprintf(w, "%s%s%s\n", promptRed, err.Error(), promptReset); werr != nil {
					return "", fmt.Errorf("components: prompt: %w", werr)
				}

				value = line
				continue
			}
		}

		return line, nil
	}
}

func confirm(r io.Reader, w io.Writer, opts ConfirmOpts) (bool, error) {
	hint := "(s/N)"

	if opts.Default {
		hint = "(S/n)"
	}

	for {
		if _, err := fmt.Fprintf(w, "%s %s: ", opts.Question, hint); err != nil {
			return false, fmt.Errorf("components: prompt: %w", err)
		}

		line, err := readLine(r)

		if err != nil {
			return false, fmt.Errorf("components: prompt: %w", err)
		}

		switch strings.ToLower(strings.TrimSpace(line)) {
		case "":
			return opts.Default, nil
		case "s", "sim", "y", "yes":
			return true, nil
		case "n", "nao", "no":
			return false, nil
		}

		if _, werr := fmt.Fprintf(w, "%sresponda s ou n%s\n", promptRed, promptReset); werr != nil {
			return false, fmt.Errorf("components: prompt: %w", werr)
		}
	}
}

func readLine(r io.Reader) (string, error) {
	var line strings.Builder
	buf := make([]byte, 1)

	for {
		n, err := r.Read(buf)

		if n > 0 {
			switch buf[0] {
			case '\n':
				return line.String(), nil
			case '\r':
			default:
				line.WriteByte(buf[0])
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				if line.Len() > 0 {
					return line.String(), nil
				}

				return "", io.ErrUnexpectedEOF
			}

			return "", err
		}
	}
}
