package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestBuiltBinary checks the command's exit codes and output on the built binary (L-W6-14).
func TestBuiltBinary(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "bootstrap-study")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	out := filepath.Join(dir, "study.tsv")
	cases := []struct {
		args []string
		code int
		want string
	}{
		{[]string{"-out", out}, 0, "----- Run Complete -----"},
		{[]string{"-B", "1"}, 2, "-B must be between 2 and 10000"},
		{[]string{"-B", "20000"}, 2, "-B must be between 2 and 10000"},
		{[]string{"-bogus"}, 2, "flag provided but not defined"},
		{[]string{"extra"}, 2, "with no other arguments"},
	}
	for _, tc := range cases {
		var stdout, stderr bytes.Buffer
		command := exec.Command(binary, tc.args...)
		command.Stdout, command.Stderr = &stdout, &stderr
		code := 0
		var exitError *exec.ExitError
		if err := command.Run(); errors.As(err, &exitError) {
			code = exitError.ExitCode()
		} else if err != nil {
			t.Fatal(err)
		}
		if code != tc.code || !strings.Contains(stdout.String()+stderr.String(), tc.want) {
			t.Errorf("%v: exit %d (want %d), output lacks %q", tc.args, code, tc.code, tc.want)
		}
	}
	table, err := os.ReadFile(out)
	if err != nil || strings.Count(string(table), "\n") != 5 {
		t.Fatalf("table: %v, %d lines (want header + 4)", err, strings.Count(string(table), "\n"))
	}
}
