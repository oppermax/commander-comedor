package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/mattn/go-isatty"
)

// withPager builds output via build, then either prints it directly (when
// stdout isn't a terminal, e.g. piped or redirected) or pipes it through a
// pager ($PAGER, falling back to "less -R" which preserves color codes).
func withPager(build func(w io.Writer)) {
	var buf bytes.Buffer
	build(&buf)

	if !isatty.IsTerminal(os.Stdout.Fd()) {
		os.Stdout.Write(buf.Bytes())
		return
	}

	pagerCmd := os.Getenv("PAGER")
	if pagerCmd == "" {
		pagerCmd = "less -R"
	}

	fields := strings.Fields(pagerCmd)
	if len(fields) == 0 {
		os.Stdout.Write(buf.Bytes())
		return
	}

	cmd := exec.Command(fields[0], fields[1:]...)
	cmd.Stdin = &buf
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Pager unavailable or failed: fall back to plain output.
		os.Stdout.Write(buf.Bytes())
	}
}
