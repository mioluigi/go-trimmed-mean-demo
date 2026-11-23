// Command tables turns the saved results into the README's generated blocks (SFEDD §5.3: every number
// in the README is generated, never typed).
//
// It reads results/comparison.tsv, results/benchmarks.txt and results/machine.txt, writes every block to
// results/tables.md, and replaces each block in README.md between the markers
//
//	<!-- BEGIN generated:NAME --> … <!-- END generated:NAME -->
//
// With -check it changes nothing and exits 1 if README.md is not what it would write.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	results := flag.String("results", "results", "folder of saved results")
	readme := flag.String("readme", "README.md", "README whose generated blocks are replaced")
	check := flag.Bool("check", false, "only check that the README is up to date")
	bootstrapReference := flag.String("bootstrap-reference", "testdata/bootstrap-reference/standard-errors.tsv", "R's standard errors for the bootstrap study")
	flag.Parse()
	if err := run(*results, *readme, *bootstrapReference, *check); err != nil {
		fmt.Fprintln(os.Stderr, "tables:", err)
		os.Exit(1)
	}
}

func run(resultsDir, readmePath, bootstrapReference string, check bool) error {
	blocks, err := buildBlocks(resultsDir, bootstrapReference)
	if err != nil {
		return err
	}
	readme, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}
	updated, err := insertBlocks(string(readme), blocks)
	if err != nil {
		return err
	}
	if check {
		if updated != string(readme) {
			return fmt.Errorf("%s is out of date with %s; run scripts/readme.sh", readmePath, resultsDir)
		}
		fmt.Println("tables: README is up to date")
		return nil
	}
	if err := os.WriteFile(filepath.Join(resultsDir, "tables.md"), []byte(allBlocks(blocks)), 0o644); err != nil {
		return err
	}
	return os.WriteFile(readmePath, []byte(updated), 0o644)
}

// buildBlocks renders every generated block from the saved results.
func buildBlocks(resultsDir, bootstrapReference string) (map[string]string, error) {
	comparisons, err := readComparisons(filepath.Join(resultsDir, "comparison.tsv"))
	if err != nil {
		return nil, err
	}
	benchmarks, err := readBenchmarks(filepath.Join(resultsDir, "benchmarks.txt"))
	if err != nil {
		return nil, err
	}
	machine, err := os.ReadFile(filepath.Join(resultsDir, "machine.txt"))
	if err != nil {
		return nil, err
	}
	bootstrap, err := bootstrapBlock(filepath.Join(resultsDir, "bootstrap-study.tsv"), bootstrapReference)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"headline":   renderHeadline(comparisons, benchmarks),
		"comparison": renderComparison(comparisons),
		"agreement":  renderAgreement(comparisons),
		"benchmarks": renderBenchmarks(benchmarks),
		"machine":    "```\n" + strings.TrimSpace(string(machine)) + "\n```\n",
		"bootstrap":  bootstrap,
	}, nil
}

// insertBlocks replaces the text between each pair of markers; a README marker with no block, or a
// block with no marker, is an error, so nothing is silently left stale.
func insertBlocks(readme string, blocks map[string]string) (string, error) {
	for name, content := range blocks {
		begin := "<!-- BEGIN generated:" + name + " -->"
		end := "<!-- END generated:" + name + " -->"
		start, finish := strings.Index(readme, begin), strings.Index(readme, end)
		if start < 0 || finish < start {
			return "", fmt.Errorf("README has no markers for block %q", name)
		}
		readme = readme[:start+len(begin)] + "\n" + content + readme[finish:]
	}
	if strings.Count(readme, "<!-- BEGIN generated:") != len(blocks) {
		return "", fmt.Errorf("README has a generated block this tool does not produce")
	}
	return readme, nil
}

func allBlocks(blocks map[string]string) string {
	names := make([]string, 0, len(blocks))
	for name := range blocks {
		names = append(names, name)
	}
	sort.Strings(names)
	var text strings.Builder
	text.WriteString("# Generated tables (cmd/tables; do not edit)\n")
	for _, name := range names {
		fmt.Fprintf(&text, "\n## %s\n\n%s", name, blocks[name])
	}
	return text.String()
}

// bootstrapBlock renders the bootstrap study (spec 002) from Go's results and R's reference.
func bootstrapBlock(goPath, rPath string) (string, error) {
	goRows, err := readBootstrap(goPath)
	if err != nil {
		return "", err
	}
	rRows, err := readBootstrap(rPath)
	if err != nil {
		return "", err
	}
	return renderBootstrap(goRows, rRows)
}
