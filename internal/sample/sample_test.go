package sample

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func writeFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "values.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestReadIntegers(t *testing.T) {
	got, err := Read(writeFile(t, "value\n3\n-1\n\n42\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsInteger() || !slices.Equal(got.Integers, []int{3, -1, 42}) || got.Len() != 3 {
		t.Fatalf("got %+v, want integers 3 -1 42", got)
	}
}

func TestReadFloatsWithoutHeader(t *testing.T) {
	got, err := Read(writeFile(t, "1.5\n2\n-0.25\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got.IsInteger() || !slices.Equal(got.Floats, []float64{1.5, 2, -0.25}) {
		t.Fatalf("got %+v, want floats 1.5 2 -0.25", got)
	}
}

func TestReadNameIsFileStem(t *testing.T) {
	got, err := Read(writeFile(t, "1\n"))
	if err != nil || got.Name != "values" {
		t.Fatalf("got name %q, err %v; want values", got.Name, err)
	}
}

func TestReadErrors(t *testing.T) {
	cases := []struct {
		name, content, wantText string
		wantErr                 error
	}{
		{"empty", "", "no numbers in", ErrNoNumbers},
		{"header only", "value\n\n", "no numbers in", ErrNoNumbers},
		{"bad line", "value\n1\n12abc\n", `line 3: "12abc" is not a number`, ErrNotNumber},
	}
	for _, tc := range cases {
		_, err := Read(writeFile(t, tc.content))
		if !errors.Is(err, tc.wantErr) || !strings.Contains(err.Error(), tc.wantText) {
			t.Errorf("%s: got %v, want %v containing %q", tc.name, err, tc.wantErr, tc.wantText)
		}
	}
	if _, err := Read(filepath.Join(t.TempDir(), "missing.csv")); err == nil {
		t.Error("missing file: got no error")
	}
}
