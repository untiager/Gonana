package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFixEmptyLines(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
		numFixes int
	}{
		{
			name:     "Remove leading empty lines",
			input:    []string{"", "", "int main() {", "}"},
			expected: []string{"int main() {", "}"},
			numFixes: 2,
		},
		{
			name:     "Remove trailing empty lines",
			input:    []string{"int main() {", "}", "", ""},
			expected: []string{"int main() {", "}"},
			numFixes: 2,
		},
		{
			name:     "Remove consecutive empty lines",
			input:    []string{"int x;", "", "", "int y;"},
			expected: []string{"int x;", "", "int y;"},
			numFixes: 1,
		},
		{
			name:     "Keep single empty lines",
			input:    []string{"int x;", "", "int y;"},
			expected: []string{"int x;", "", "int y;"},
			numFixes: 0,
		},
		{
			name:     "Clean file",
			input:    []string{"int x;", "int y;"},
			expected: []string{"int x;", "int y;"},
			numFixes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixer := NewFixer(nil, true)
			result := &FixResult{Fixes: make([]Fix, 0)}
			fixed := fixer.fixEmptyLines(tt.input, result)

			if len(fixed) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(fixed))
			}

			for i := range fixed {
				if i < len(tt.expected) && fixed[i] != tt.expected[i] {
					t.Errorf("Line %d: expected %q, got %q", i, tt.expected[i], fixed[i])
				}
			}

			if len(result.Fixes) != tt.numFixes {
				t.Errorf("Expected %d fixes, got %d", tt.numFixes, len(result.Fixes))
			}
		})
	}
}

func TestFixIndentation(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
		numFixes int
	}{
		{
			name:     "Replace 4 spaces with tab",
			input:    []string{"    int x;"},
			expected: []string{"\tint x;"},
			numFixes: 1,
		},
		{
			name:     "Replace 8 spaces with 2 tabs",
			input:    []string{"        int x;"},
			expected: []string{"\t\tint x;"},
			numFixes: 1,
		},
		{
			name:     "Keep tabs",
			input:    []string{"\tint x;"},
			expected: []string{"\tint x;"},
			numFixes: 0,
		},
		{
			name:     "Mixed spaces (5 spaces = 1 tab + 1 space)",
			input:    []string{"     int x;"},
			expected: []string{"\t int x;"},
			numFixes: 1,
		},
		{
			name:     "No indentation",
			input:    []string{"int x;"},
			expected: []string{"int x;"},
			numFixes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixer := NewFixer(nil, true)
			result := &FixResult{Fixes: make([]Fix, 0)}
			fixed := fixer.fixIndentation(tt.input, result)

			if len(fixed) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(fixed))
			}

			for i := range fixed {
				if i < len(tt.expected) && fixed[i] != tt.expected[i] {
					t.Errorf("Line %d: expected %q, got %q", i, tt.expected[i], fixed[i])
				}
			}

			if len(result.Fixes) != tt.numFixes {
				t.Errorf("Expected %d fixes, got %d", tt.numFixes, len(result.Fixes))
			}
		})
	}
}

func TestFixMultipleVariableDeclarations(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
		numFixes int
	}{
		{
			name:     "Split two variables",
			input:    []string{"int x, y;"},
			expected: []string{"int x;", "int y;"},
			numFixes: 1,
		},
		{
			name:     "Split three variables",
			input:    []string{"int a, b, c;"},
			expected: []string{"int a;", "int b;", "int c;"},
			numFixes: 1,
		},
		{
			name:     "Keep single variable",
			input:    []string{"int x;"},
			expected: []string{"int x;"},
			numFixes: 0,
		},
		{
			name:     "Preserve indentation",
			input:    []string{"\tint x, y;"},
			expected: []string{"\tint x;", "\tint y;"},
			numFixes: 1,
		},
		{
			name:     "Skip for loops",
			input:    []string{"for (int i = 0, j = 0; i < 10; i++)"},
			expected: []string{"for (int i = 0, j = 0; i < 10; i++)"},
			numFixes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixer := NewFixer(nil, true)
			result := &FixResult{Fixes: make([]Fix, 0)}
			fixed := fixer.fixMultipleVariableDeclarations(tt.input, result)

			if len(fixed) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(fixed))
			}

			for i := range fixed {
				if i < len(tt.expected) && fixed[i] != tt.expected[i] {
					t.Errorf("Line %d: expected %q, got %q", i, tt.expected[i], fixed[i])
				}
			}

			if len(result.Fixes) != tt.numFixes {
				t.Errorf("Expected %d fixes, got %d", tt.numFixes, len(result.Fixes))
			}
		})
	}
}

func TestFixCommentFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
		numFixes int
	}{
		{
			name:     "Convert simple comment",
			input:    []string{"// This is a comment"},
			expected: []string{"/* This is a comment */"},
			numFixes: 1,
		},
		{
			name:     "Convert inline comment",
			input:    []string{"int x; // Variable"},
			expected: []string{"int x; /* Variable */"},
			numFixes: 1,
		},
		{
			name:     "Keep block comments",
			input:    []string{"/* This is a comment */"},
			expected: []string{"/* This is a comment */"},
			numFixes: 0,
		},
		{
			name:     "Empty comment",
			input:    []string{"int x; //"},
			expected: []string{"int x;"},
			numFixes: 1,
		},
		{
			name:     "Multiple lines",
			input:    []string{"// Comment 1", "int x;", "// Comment 2"},
			expected: []string{"/* Comment 1 */", "int x;", "/* Comment 2 */"},
			numFixes: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixer := NewFixer(nil, true)
			result := &FixResult{Fixes: make([]Fix, 0)}
			fixed := fixer.fixCommentFormat(tt.input, result)

			if len(fixed) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(fixed))
			}

			for i := range fixed {
				if i < len(tt.expected) && fixed[i] != tt.expected[i] {
					t.Errorf("Line %d: expected %q, got %q", i, tt.expected[i], fixed[i])
				}
			}

			if len(result.Fixes) != tt.numFixes {
				t.Errorf("Expected %d fixes, got %d", tt.numFixes, len(result.Fixes))
			}
		})
	}
}

func TestFixForLoopDeclarations(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
		numFixes int
	}{
		{
			name:     "Extract int declaration",
			input:    []string{"for (int i = 0; i < 10; i++)"},
			expected: []string{"int i;", "", "for (i = 0; i < 10; i++)"},
			numFixes: 1,
		},
		{
			name:     "Keep normal for loop",
			input:    []string{"for (i = 0; i < 10; i++)"},
			expected: []string{"for (i = 0; i < 10; i++)"},
			numFixes: 0,
		},
		{
			name:     "Preserve indentation",
			input:    []string{"\tfor (int i = 0; i < 10; i++)"},
			expected: []string{"\tint i;", "", "\tfor (i = 0; i < 10; i++)"},
			numFixes: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixer := NewFixer(nil, true)
			result := &FixResult{Fixes: make([]Fix, 0)}
			fixed := fixer.fixForLoopDeclarations(tt.input, result)

			if len(fixed) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(fixed))
				t.Logf("Expected: %v", tt.expected)
				t.Logf("Got: %v", fixed)
			}

			for i := range fixed {
				if i < len(tt.expected) && fixed[i] != tt.expected[i] {
					t.Errorf("Line %d: expected %q, got %q", i, tt.expected[i], fixed[i])
				}
			}

			if len(result.Fixes) != tt.numFixes {
				t.Errorf("Expected %d fixes, got %d", tt.numFixes, len(result.Fixes))
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HelloWorld", "hello_world"},
		{"myFileName", "my_file_name"},
		{"already_snake", "already_snake"},
		{"TestCase", "test_case"},
		{"simple", "simple"},
		{"ABCTest", "a_b_c_test"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestShouldFixFilename(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"my_file.c", false},
		{"MyFile.c", true},
		{"camelCase.c", true},
		{"snake_case.h", false},
		{"test_123.c", false},
		{"Test.c", true},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			fixer := NewFixer(nil, true)
			result := fixer.shouldFixFilename(tt.filename)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFixFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"path/to/MyFile.c", "path/to/my_file.c"},
		{"TestCase.h", "test_case.h"},
		{"already_snake.c", "already_snake.c"},
		{"/abs/path/CamelCase.c", "/abs/path/camel_case.c"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			fixer := NewFixer(nil, true)
			result := fixer.fixFilename(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFixFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.c")

	content := `

int x, y;
    int z;
// This is a comment
for (int i = 0; i < 10; i++) {
    printf("test");
}

`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test dry run
	t.Run("Dry run", func(t *testing.T) {
		analyzer := NewAnalyzer(1)
		fixer := NewFixer(analyzer, true)

		result, err := fixer.FixFile(testFile)
		if err != nil {
			t.Fatal(err)
		}

		if result.ModifiedContent {
			t.Error("Expected no modification in dry run")
		}

		if len(result.Fixes) == 0 {
			t.Error("Expected some fixes to be identified")
		}

		// Verify file wasn't actually modified
		readBack, _ := os.ReadFile(testFile)
		if string(readBack) != content {
			t.Error("File was modified during dry run")
		}
	})

	// Test actual fix
	t.Run("Actual fix", func(t *testing.T) {
		analyzer := NewAnalyzer(1)
		fixer := NewFixer(analyzer, false)

		result, err := fixer.FixFile(testFile)
		if err != nil {
			t.Fatal(err)
		}

		if !result.ModifiedContent {
			t.Error("Expected file to be modified")
		}

		if len(result.Fixes) == 0 {
			t.Error("Expected some fixes to be applied")
		}

		// Verify file was actually modified
		readBack, _ := os.ReadFile(testFile)
		if string(readBack) == content {
			t.Error("File was not modified")
		}

		// Check some expected fixes
		fixed := string(readBack)
		if strings.Contains(fixed, "//") {
			t.Error("Expected // comments to be converted")
		}
		if strings.Contains(fixed, "int x, y;") {
			t.Error("Expected multiple declarations to be split")
		}
		if !strings.HasPrefix(fixed, "int") {
			t.Error("Expected leading empty lines to be removed")
		}
		if strings.HasSuffix(fixed, "\n\n") {
			t.Error("Expected trailing empty lines to be removed")
		}
	})
}

func TestFixFile_InvalidFile(t *testing.T) {
	analyzer := NewAnalyzer(1)
	fixer := NewFixer(analyzer, true)

	_, err := fixer.FixFile("/nonexistent/file.c")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestCollectCFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure
	os.WriteFile(filepath.Join(tmpDir, "test.c"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test.h"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte(""), 0644)

	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "sub.c"), []byte(""), 0644)

	files, err := collectCFiles(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 3 {
		t.Errorf("Expected 3 C files, got %d", len(files))
	}

	// Test single file
	files, err = collectCFiles(filepath.Join(tmpDir, "test.c"))
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}
}
