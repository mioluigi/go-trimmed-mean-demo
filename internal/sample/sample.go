// Package sample reads a column of numbers from a text file, as integers when every value is a whole
// number and as floats otherwise.
package sample

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Errors returned by Read; the message also names the file and, for ErrNotNumber, the line.
var (
	// ErrNoNumbers means the file holds no values (it may hold only the header).
	ErrNoNumbers = errors.New("no numbers")
	// ErrNotNumber means a line is neither a number nor the header.
	ErrNotNumber = errors.New("not a number")
)

// header is the optional first line of a sample file.
const header = "value"

// Sample is one column of numbers. Exactly one of Integers and Floats holds the values.
type Sample struct {
	Name     string // the file name without directory and extension, e.g. "floats-100"
	Integers []int
	Floats   []float64
}

// IsInteger reports whether the sample was read as integers.
func (s Sample) IsInteger() bool { return s.Integers != nil }

// Len returns the number of values.
func (s Sample) Len() int { return len(s.Integers) + len(s.Floats) }

// Read reads one number per line from path. Blank lines and a first line "value" are skipped. If every
// value is a whole number written without a decimal point the sample is integer, otherwise float.
func Read(path string) (Sample, error) {
	texts, err := readValueLines(path)
	if err != nil {
		return Sample{}, err
	}
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if integers, ok := parseIntegers(texts); ok {
		return Sample{Name: name, Integers: integers}, nil
	}
	floats, err := parseFloats(path, texts)
	if err != nil {
		return Sample{}, err
	}
	return Sample{Name: name, Floats: floats}, nil
}

// valueLine is a non-blank line and its 1-based line number, kept for error messages.
type valueLine struct {
	number int
	text   string
}

// readValueLines returns the non-blank lines after the optional header.
func readValueLines(path string) ([]valueLine, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []valueLine
	scanner := bufio.NewScanner(file)
	for number := 1; scanner.Scan(); number++ {
		text := strings.TrimSpace(scanner.Text())
		isHeader := number == 1 && text == header
		if text != "" && !isHeader {
			lines = append(lines, valueLine{number, text})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("%w in %s", ErrNoNumbers, path)
	}
	return lines, nil
}

// parseIntegers parses every line as an integer, or reports false if any line is not one.
func parseIntegers(lines []valueLine) ([]int, bool) {
	integers := make([]int, len(lines))
	for i, line := range lines {
		value, err := strconv.Atoi(line.text)
		if err != nil {
			return nil, false
		}
		integers[i] = value
	}
	return integers, true
}

// parseFloats parses every line as a float64, naming the first line that is not a number.
func parseFloats(path string, lines []valueLine) ([]float64, error) {
	floats := make([]float64, len(lines))
	for i, line := range lines {
		value, err := strconv.ParseFloat(line.text, 64)
		if err != nil {
			return nil, fmt.Errorf("%s line %d: %q is %w", path, line.number, line.text, ErrNotNumber)
		}
		floats[i] = value
	}
	return floats, nil
}
