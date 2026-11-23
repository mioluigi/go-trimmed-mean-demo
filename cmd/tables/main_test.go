package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readmeWithMarkers is a README holding every block's markers around stale text.
func readmeWithMarkers() string {
	var text strings.Builder
	text.WriteString("# Title\n")
	for _, name := range []string{"headline", "comparison", "agreement", "benchmarks", "machine", "bootstrap"} {
		text.WriteString("<!-- BEGIN generated:" + name + " -->\nstale\n<!-- END generated:" + name + " -->\n")
	}
	return text.String()
}

// fixture copies testdata (a real comparison.tsv from the program, hand-written benchmark lines) and a
// README into a temporary folder.
func fixture(t *testing.T, readme string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"comparison.tsv", "benchmarks.txt", "machine.txt", "bootstrap-study.tsv", "bootstrap-reference.tsv"} {
		content, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	readmePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readme), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, readmePath
}

func TestRunFillsBlocksAndCheckDetectsStale(t *testing.T) {
	dir, readmePath := fixture(t, readmeWithMarkers())
	if err := run(dir, readmePath, filepath.Join(dir, "bootstrap-reference.tsv"), true); err == nil {
		t.Fatal("check passed on a stale README")
	}
	if err := run(dir, readmePath, filepath.Join(dir, "bootstrap-reference.tsv"), false); err != nil {
		t.Fatal(err)
	}
	readme, _ := os.ReadFile(readmePath)
	for _, want := range []string{
		"**84 of 84** comparisons", "exactly** on **3 of 3** integer samples", "| integers-100 | 100 | integer | 50.81111111111111 |",
		"| floats/n=100 | 3 | 1.60 µs |", "Slowest benchmark (`ints/n=1000000`): **62.00 ms**", "machine: fixture",
		"| 100 | 1.00 | 1.41 | 1.42 | 1.11 | **1.02** |", "smallest standard error of the three, by both measures, at n = **100, 225, 400**", "largest relative difference of **1e-12**",
	} {
		if !strings.Contains(string(readme), want) {
			t.Errorf("README lacks %q", want)
		}
	}
	if strings.Contains(string(readme), "stale") {
		t.Error("stale text left in the README")
	}
	if err := run(dir, readmePath, filepath.Join(dir, "bootstrap-reference.tsv"), true); err != nil {
		t.Fatalf("check failed right after generating: %v", err)
	}
}

func TestRunRejectsMissingMarkers(t *testing.T) {
	dir, readmePath := fixture(t, "# Title without markers\n")
	if err := run(dir, readmePath, filepath.Join(dir, "bootstrap-reference.tsv"), false); err == nil {
		t.Fatal("a README without markers was accepted")
	}
}
