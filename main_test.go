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

// binary is the program built once by TestMain; exit codes are tested on it, never through go run
// (which reports 1 for every failure).
var binary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "trimmed-mean-demo")
	if err != nil {
		panic(err)
	}
	binary = filepath.Join(dir, "go-trimmed-mean-demo")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		panic("build failed: " + string(output))
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// runBinary runs the built program from the repository root and returns its exit code and output.
func runBinary(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	command := exec.Command(binary, args...)
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	var exitError *exec.ExitError
	switch {
	case err == nil:
		return 0, stdout.String(), stderr.String()
	case errors.As(err, &exitError):
		return exitError.ExitCode(), stdout.String(), stderr.String()
	default:
		t.Fatal(err)
		return 0, "", ""
	}
}

func writeInput(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "input.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestComparisonMode is AC-9: no arguments, every sample compared, plain-language result, full table.
func TestComparisonMode(t *testing.T) {
	out := filepath.Join(t.TempDir(), "comparison.tsv")
	code, stdout, stderr := runBinary(t, "-out", out)
	if code != exitOK {
		t.Fatalf("exit %d, stderr:\n%s", code, stderr)
	}
	for _, name := range []string{"integers-100", "integers-101", "integers-1000", "floats-100", "floats-101", "floats-1000"} {
		if !strings.Contains(stdout, name) {
			t.Errorf("stdout does not mention %s", name)
		}
	}
	for _, want := range []string{"mean(x, trim = 0.05)", "PASS", "84 of 84 comparisons with R agree"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout lacks %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "level=INFO") || !strings.Contains(stderr, "level=INFO") {
		t.Error("logs must go to standard error only")
	}
	table, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.Count(string(table), "\n"); lines != 85 {
		t.Errorf("table has %d lines, want 85 (header + 84 rows)", lines)
	}
}

// TestComparisonFailure plants a wrong R value: the run must report FAIL and exit 1.
func TestComparisonFailure(t *testing.T) {
	original, err := os.ReadFile("testdata/r-reference/trimmed-means.tsv")
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(string(original), "50.81111111111111", "50.81111111111112", 1)
	reference := writeInput(t, tampered)
	outDir := t.TempDir()
	code, stdout, stderr := runBinary(t, "-reference", reference, "-out", filepath.Join(outDir, "c.tsv"))
	if code != exitFailure || !strings.Contains(stdout, "FAIL") || !strings.Contains(stderr, "1 comparisons with R failed") {
		t.Fatalf("exit %d, stdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	// Review R2: a failed run leaves the table it wrote and no temporary file.
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "c.tsv" {
		t.Fatalf("output folder holds %v, want only c.tsv", entries)
	}
}

func TestOneFileMode(t *testing.T) {
	cases := []struct {
		args []string
		want []string
	}{
		{[]string{"-data", "data/integers-100.csv", "-trim", "0.05"},
			[]string{"100 integers", "(5 and 5 values dropped)", "trimmed mean: 50.81111111111111"}},
		{[]string{"-data", "data/integers-100.csv"},
			[]string{"trimmed mean: 50.81111111111111"}},
		{[]string{"-data", "data/floats-100.csv", "-trim", "0.05,0.10"},
			[]string{"100 floats", "0.05 from the bottom and 0.1 from the top (5 and 10 values dropped)"}},
	}
	for _, tc := range cases {
		code, stdout, stderr := runBinary(t, tc.args...)
		if code != exitOK {
			t.Errorf("%v: exit %d, stderr %s", tc.args, code, stderr)
			continue
		}
		for _, want := range tc.want {
			if !strings.Contains(stdout, want) {
				t.Errorf("%v: stdout lacks %q:\n%s", tc.args, want, stdout)
			}
		}
	}
}

// TestInputErrors is AC-10: each boundary of spec §8.1 gives its message and exit code.
func TestInputErrors(t *testing.T) {
	notNumber := writeInput(t, "value\n1\n12abc\n")
	empty := writeInput(t, "value\n")
	withNaN := writeInput(t, "1\nNaN\n3\n")
	infinities := writeInput(t, "-Inf\n1\nInf\n")
	cases := []struct {
		name string
		args []string
		code int
		want string
	}{
		{"proportion too large", []string{"-data", "data/floats-100.csv", "-trim", "0.6"}, exitUsage, "trim proportion must be in [0, 0.5): got 0.6"},
		{"negative proportion", []string{"-data", "data/floats-100.csv", "-trim", "-0.1"}, exitUsage, "got -0.1"},
		{"three proportions", []string{"-data", "data/floats-100.csv", "-trim", "0.1,0.1,0.1"}, exitUsage, "at most two trim proportions"},
		{"malformed trim", []string{"-data", "data/floats-100.csv", "-trim", "abc"}, exitUsage, `-trim must be one proportion`},
		{"trim without data", []string{"-trim", "0.1"}, exitUsage, "-trim needs -data FILE"},
		{"missing file", []string{"-data", "no-such-file.csv"}, exitUsage, "no such file"},
		{"not a number", []string{"-data", notNumber}, exitUsage, `line 3: "12abc" is not a number`},
		{"no numbers", []string{"-data", empty}, exitUsage, "no numbers in"},
		{"NaN value", []string{"-data", withNaN}, exitUsage, "value is NaN: at index 1"},
		{"infinite result", []string{"-data", infinities, "-trim", "0"}, exitFailure, "mean is not finite"},
		{"unknown flag", []string{"-bogus"}, exitUsage, "usage:"},
		{"stray argument", []string{"extra"}, exitUsage, `unexpected argument "extra"`},
	}
	for _, tc := range cases {
		code, _, stderr := runBinary(t, tc.args...)
		if code != tc.code || !strings.Contains(stderr, tc.want) {
			t.Errorf("%s: exit %d (want %d), stderr %q (want %q)", tc.name, code, tc.code, stderr, tc.want)
		}
	}
}

// TestProfiles checks that -cpuprofile and -memprofile write non-empty files.
func TestProfiles(t *testing.T) {
	dir := t.TempDir()
	cpu, mem := filepath.Join(dir, "cpu.prof"), filepath.Join(dir, "mem.prof")
	code, _, stderr := runBinary(t, "-data", "data/floats-1000.csv", "-cpuprofile", cpu, "-memprofile", mem)
	if code != exitOK {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	for _, path := range []string{cpu, mem} {
		if info, err := os.Stat(path); err != nil || info.Size() == 0 {
			t.Errorf("%s missing or empty", path)
		}
	}
}
