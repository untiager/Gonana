// main_test.go
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test helper functions
func TestIsSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid snake_case", "my_function", true},
		{"valid single word", "function", true},
		{"invalid camelCase", "myFunction", false},
		{"invalid PascalCase", "MyFunction", false},
		{"invalid uppercase", "MY_FUNCTION", false},
		{"invalid leading underscore", "_function", false},
		{"invalid trailing underscore", "function_", false},
		{"empty string", "", false},
		{"multiple underscores", "my_long_function_name", true},
		{"with numbers", "my_func_2", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("isSnakeCase(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsScreamingSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid SCREAMING_SNAKE_CASE", "MY_MACRO", true},
		{"valid single word", "MACRO", true},
		{"invalid lowercase", "my_macro", false},
		{"invalid mixed case", "My_Macro", false},
		{"invalid leading underscore", "_MACRO", false},
		{"invalid trailing underscore", "MACRO_", false},
		{"empty string", "", false},
		{"multiple underscores", "MY_LONG_MACRO_NAME", true},
		{"with numbers", "MY_MACRO_2", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isScreamingSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("isScreamingSnakeCase(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractFunctions(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected int
		funcName string
	}{
		{
			name: "simple function",
			lines: []string{
				"int my_function(void)",
				"{",
				"    return 0;",
				"}",
			},
			expected: 1,
			funcName: "my_function",
		},
		{
			name: "function with parameters",
			lines: []string{
				"int add_numbers(int a, int b, int c)",
				"{",
				"    return a + b + c;",
				"}",
			},
			expected: 1,
			funcName: "add_numbers",
		},
		{
			name: "multiple functions",
			lines: []string{
				"int func1(void) {",
				"    return 1;",
				"}",
				"",
				"void func2(void) {",
				"    return;",
				"}",
			},
			expected: 2,
			funcName: "func1",
		},
		{
			name: "no functions",
			lines: []string{
				"#include <stdio.h>",
				"int x = 5;",
			},
			expected: 0,
			funcName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			functions := extractFunctions(tt.lines)
			if len(functions) != tt.expected {
				t.Errorf("extractFunctions() returned %d functions, want %d", len(functions), tt.expected)
			}
			if tt.expected > 0 && len(functions) > 0 {
				if functions[0].Name != tt.funcName {
					t.Errorf("extractFunctions() first function name = %q, want %q", functions[0].Name, tt.funcName)
				}
			}
		})
	}
}

// Test rule checking functions
func TestCheckLineLength(t *testing.T) {
	analysis := &FileAnalysis{
		Lines: []string{
			"short line",
			"this is a very long line that exceeds the maximum allowed length of 80 characters and should be reported",
			"another short line",
		},
	}

	violations := checkLineLength(analysis, "test.c", 0)
	if len(violations) != 1 {
		t.Errorf("checkLineLength() found %d violations, want 1", len(violations))
	}
	if len(violations) > 0 && violations[0].Rule != "C-L1" {
		t.Errorf("checkLineLength() violation rule = %q, want %q", violations[0].Rule, "C-L1")
	}
	if len(violations) > 0 && violations[0].Line != 2 {
		t.Errorf("checkLineLength() violation line = %d, want 2", violations[0].Line)
	}
}

func TestCheckEmptyLines(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected int
	}{
		{
			name: "empty line at start",
			lines: []string{
				"",
				"int x;",
			},
			expected: 1,
		},
		{
			name: "empty line at end",
			lines: []string{
				"int x;",
				"",
			},
			expected: 1,
		},
		{
			name: "consecutive empty lines",
			lines: []string{
				"int x;",
				"",
				"",
				"int y;",
			},
			expected: 1,
		},
		{
			name: "no violations",
			lines: []string{
				"int x;",
				"",
				"int y;",
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := &FileAnalysis{Lines: tt.lines}
			violations := checkEmptyLines(analysis, "test.c", 0)
			if len(violations) != tt.expected {
				t.Errorf("checkEmptyLines() found %d violations, want %d", len(violations), tt.expected)
			}
		})
	}
}

func TestCheckIndentation(t *testing.T) {
	analysis := &FileAnalysis{
		Lines: []string{
			"int main(void)",
			"{",
			"\tint x;",   // tab indentation - good
			"    int y;", // space indentation - bad
			"\treturn 0;",
			"}",
		},
	}

	violations := checkIndentation(analysis, "test.c", 0)
	if len(violations) != 1 {
		t.Errorf("checkIndentation() found %d violations, want 1", len(violations))
	}
	if len(violations) > 0 && violations[0].Rule != "C-L3" {
		t.Errorf("checkIndentation() violation rule = %q, want %q", violations[0].Rule, "C-L3")
	}
}

func TestCheckVariableDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected int
	}{
		{
			name: "multiple declaration",
			lines: []string{
				"int x, y, z;",
			},
			expected: 1,
		},
		{
			name: "single declaration",
			lines: []string{
				"int x;",
				"int y;",
			},
			expected: 0,
		},
		{
			name: "for loop ignored",
			lines: []string{
				"for (int i = 0, j = 0; i < 10; i++)",
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := &FileAnalysis{Lines: tt.lines}
			violations := checkVariableDeclaration(analysis, "test.c", 0)
			if len(violations) != tt.expected {
				t.Errorf("checkVariableDeclaration() found %d violations, want %d", len(violations), tt.expected)
			}
		})
	}
}

func TestCheckFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected int
	}{
		{"valid snake_case", "my_file.c", 0},
		{"invalid camelCase", "myFile.c", 1},
		{"invalid PascalCase", "MyFile.c", 1},
		{"invalid uppercase", "MY_FILE.c", 1},
		{"valid with numbers", "file_2.c", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := &FileAnalysis{}
			violations := checkFilename(analysis, tt.filename, 0)
			if len(violations) != tt.expected {
				t.Errorf("checkFilename(%q) found %d violations, want %d", tt.filename, len(violations), tt.expected)
			}
		})
	}
}

func TestCheckFunctionNames(t *testing.T) {
	tests := []struct {
		name      string
		functions []FunctionInfo
		expected  int
	}{
		{
			name: "valid function name",
			functions: []FunctionInfo{
				{Name: "my_function", StartLine: 1},
			},
			expected: 0,
		},
		{
			name: "invalid camelCase",
			functions: []FunctionInfo{
				{Name: "myFunction", StartLine: 1},
			},
			expected: 1,
		},
		{
			name: "main function allowed",
			functions: []FunctionInfo{
				{Name: "main", StartLine: 1},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := &FileAnalysis{Functions: tt.functions}
			violations := checkFunctionNames(analysis, "test.c", 0)
			if len(violations) != tt.expected {
				t.Errorf("checkFunctionNames() found %d violations, want %d", len(violations), tt.expected)
			}
		})
	}
}

func TestCheckMacroNames(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected int
	}{
		{
			name: "valid SCREAMING_SNAKE_CASE",
			lines: []string{
				"#define MY_MACRO 42",
			},
			expected: 0,
		},
		{
			name: "invalid lowercase",
			lines: []string{
				"#define my_macro 42",
			},
			expected: 1,
		},
		{
			name: "invalid mixed case",
			lines: []string{
				"#define MyMacro 42",
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := &FileAnalysis{Lines: tt.lines}
			violations := checkMacroNames(analysis, "test.c", 0)
			if len(violations) != tt.expected {
				t.Errorf("checkMacroNames() found %d violations, want %d", len(violations), tt.expected)
			}
		})
	}
}

func TestCheckFunctionLength(t *testing.T) {
	tests := []struct {
		name      string
		functions []FunctionInfo
		expected  int
	}{
		{
			name: "short function",
			functions: []FunctionInfo{
				{Name: "short_func", StartLine: 1, EndLine: 10},
			},
			expected: 0,
		},
		{
			name: "long function",
			functions: []FunctionInfo{
				{Name: "long_func", StartLine: 1, EndLine: 30},
			},
			expected: 1,
		},
		{
			name: "exactly 25 lines",
			functions: []FunctionInfo{
				{Name: "exact_func", StartLine: 1, EndLine: 25},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := &FileAnalysis{Functions: tt.functions}
			violations := checkFunctionLength(analysis, "test.c", 0)
			if len(violations) != tt.expected {
				t.Errorf("checkFunctionLength() found %d violations, want %d", len(violations), tt.expected)
			}
		})
	}
}

func TestCheckCommentFormat(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected int
	}{
		{
			name: "valid block comment",
			lines: []string{
				"/* This is a comment */",
			},
			expected: 0,
		},
		{
			name: "invalid line comment",
			lines: []string{
				"// This is a comment",
			},
			expected: 1,
		},
		{
			name: "multiple line comments",
			lines: []string{
				"// Comment 1",
				"// Comment 2",
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := &FileAnalysis{Lines: tt.lines}
			violations := checkCommentFormat(analysis, "test.c", 0)
			if len(violations) != tt.expected {
				t.Errorf("checkCommentFormat() found %d violations, want %d", len(violations), tt.expected)
			}
		})
	}
}

func TestCheckFunctionParameters(t *testing.T) {
	tests := []struct {
		name      string
		functions []FunctionInfo
		expected  int
	}{
		{
			name: "valid parameter count",
			functions: []FunctionInfo{
				{Name: "func", StartLine: 1, ParamCount: 3},
			},
			expected: 0,
		},
		{
			name: "too many parameters",
			functions: []FunctionInfo{
				{Name: "func", StartLine: 1, ParamCount: 5},
			},
			expected: 1,
		},
		{
			name: "exactly 4 parameters",
			functions: []FunctionInfo{
				{Name: "func", StartLine: 1, ParamCount: 4},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := &FileAnalysis{Functions: tt.functions}
			violations := checkFunctionParameters(analysis, "test.c", 0)
			if len(violations) != tt.expected {
				t.Errorf("checkFunctionParameters() found %d violations, want %d", len(violations), tt.expected)
			}
		})
	}
}

func TestCheckForLoopDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected int
	}{
		{
			name: "declaration in for loop",
			lines: []string{
				"for (int i = 0; i < 10; i++)",
			},
			expected: 1,
		},
		{
			name: "no declaration in for loop",
			lines: []string{
				"int i;",
				"for (i = 0; i < 10; i++)",
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := &FileAnalysis{Lines: tt.lines}
			violations := checkForLoopDeclaration(analysis, "test.c", 0)
			if len(violations) != tt.expected {
				t.Errorf("checkForLoopDeclaration() found %d violations, want %d", len(violations), tt.expected)
			}
		})
	}
}

// Test Analyzer
func TestNewAnalyzer(t *testing.T) {
	tests := []struct {
		name          string
		level         int
		expectedRules int
	}{
		{"level 1", 1, 10}, // 10 level 1 rules
		{"level 2", 2, 15}, // 10 level 1 + 5 level 2 rules
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewAnalyzer(tt.level)
			if analyzer.level != tt.level {
				t.Errorf("NewAnalyzer(%d).level = %d, want %d", tt.level, analyzer.level, tt.level)
			}
			if len(analyzer.rules) != tt.expectedRules {
				t.Errorf("NewAnalyzer(%d) has %d rules, want %d", tt.level, len(analyzer.rules), tt.expectedRules)
			}
		})
	}
}

func TestAnalyzeFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_file.c")

	content := `int my_function(void)
{
	int x;
	return x;
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	analyzer := NewAnalyzer(1)
	result, err := analyzer.analyzeFile(testFile)
	if err != nil {
		t.Fatalf("analyzeFile() error = %v", err)
	}

	if result == nil {
		t.Fatal("analyzeFile() returned nil result")
	}

	if result.Filename != "test_file.c" {
		t.Errorf("analyzeFile() filename = %q, want %q", result.Filename, "test_file.c")
	}

	if result.LineCount != 6 {
		t.Errorf("analyzeFile() line count = %d, want 6", result.LineCount)
	}

	if result.Score > 100 || result.Score < 0 {
		t.Errorf("analyzeFile() score = %f, want between 0 and 100", result.Score)
	}
}

func TestAnalyzePath_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.c")

	content := `int x;
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	analyzer := NewAnalyzer(1)
	report, err := analyzer.AnalyzePath(testFile)
	if err != nil {
		t.Fatalf("AnalyzePath() error = %v", err)
	}

	if report.TotalFiles != 1 {
		t.Errorf("AnalyzePath() total files = %d, want 1", report.TotalFiles)
	}

	if len(report.Files) != 1 {
		t.Errorf("AnalyzePath() files count = %d, want 1", len(report.Files))
	}
}

func TestAnalyzePath_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple test files
	testFiles := []string{"file1.c", "file2.c", "file3.h"}
	for _, name := range testFiles {
		content := "int x;\n"
		err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", name, err)
		}
	}

	// Create a non-C file that should be ignored
	os.WriteFile(filepath.Join(tmpDir, "ignore.txt"), []byte("test"), 0644)

	analyzer := NewAnalyzer(1)
	report, err := analyzer.AnalyzePath(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzePath() error = %v", err)
	}

	if report.TotalFiles != 3 {
		t.Errorf("AnalyzePath() total files = %d, want 3", report.TotalFiles)
	}
}

func TestAnalyzePath_InvalidPath(t *testing.T) {
	analyzer := NewAnalyzer(1)
	_, err := analyzer.AnalyzePath("/nonexistent/path/to/file.c")
	if err == nil {
		t.Error("AnalyzePath() with invalid path should return error")
	}
}

func TestAnalyzeFile_WithViolations(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "bad_file.c")

	// Create a file with multiple violations
	content := `
int myBadFunction(void)
{
    int x, y, z;
    this is a very long line that exceeds the maximum allowed length of 80 characters and should trigger a violation
    return 0;
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	analyzer := NewAnalyzer(1)
	result, err := analyzer.analyzeFile(testFile)
	if err != nil {
		t.Fatalf("analyzeFile() error = %v", err)
	}

	if len(result.Violations) == 0 {
		t.Error("analyzeFile() expected violations but found none")
	}

	// Check that score is less than 100 due to violations
	if result.Score >= 100 {
		t.Errorf("analyzeFile() score = %f, want less than 100 due to violations", result.Score)
	}

	// Verify specific violations exist
	hasLineViolation := false
	hasIndentViolation := false
	hasMultiVarViolation := false

	for _, v := range result.Violations {
		if v.Rule == "C-L1" {
			hasLineViolation = true
		}
		if v.Rule == "C-L3" {
			hasIndentViolation = true
		}
		if v.Rule == "C-L4" {
			hasMultiVarViolation = true
		}
	}

	if !hasLineViolation {
		t.Error("Expected C-L1 (line length) violation")
	}
	if !hasIndentViolation {
		t.Error("Expected C-L3 (indentation) violation")
	}
	if !hasMultiVarViolation {
		t.Error("Expected C-L4 (multiple variable declaration) violation")
	}
}

func TestAnalyzeFile_CleanFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "clean_file.c")

	// Create a clean file with no violations (no trailing newline)
	content := "int my_function(void)\n{\n\tint x;\n\tint y;\n\n\tx = 42;\n\ty = x + 1;\n\treturn y;\n}"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	analyzer := NewAnalyzer(1)
	result, err := analyzer.analyzeFile(testFile)
	if err != nil {
		t.Fatalf("analyzeFile() error = %v", err)
	}

	if len(result.Violations) != 0 {
		t.Errorf("analyzeFile() found %d violations in clean file, want 0", len(result.Violations))
		for _, v := range result.Violations {
			t.Logf("  Violation: %s - %s (line %d)", v.Rule, v.Message, v.Line)
		}
	}

	if result.Score != 100 {
		t.Errorf("analyzeFile() score = %f, want 100 for clean file", result.Score)
	}
}

func TestReport_Calculations(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with known violations (no trailing newlines)
	file1 := filepath.Join(tmpDir, "file1.c")
	os.WriteFile(file1, []byte("int x;"), 0644)

	file2 := filepath.Join(tmpDir, "file2.c")
	os.WriteFile(file2, []byte("    int y;"), 0644) // space indentation violation

	analyzer := NewAnalyzer(1)
	report, err := analyzer.AnalyzePath(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzePath() error = %v", err)
	}

	if report.TotalFiles != 2 {
		t.Errorf("Report total files = %d, want 2", report.TotalFiles)
	}

	if report.CleanFiles != 1 {
		t.Errorf("Report clean files = %d, want 1", report.CleanFiles)
	}

	if report.TotalViolations == 0 {
		t.Error("Report should have violations")
	}

	if report.TotalScore == 0 {
		t.Error("Report should have calculated total score")
	}
}

func TestExtractFunctions_WithParameters(t *testing.T) {
	lines := []string{
		"int add(int a, int b, int c, int d)",
		"{",
		"	return a + b + c + d;",
		"}",
	}

	functions := extractFunctions(lines)
	if len(functions) != 1 {
		t.Fatalf("extractFunctions() returned %d functions, want 1", len(functions))
	}

	if functions[0].ParamCount != 4 {
		t.Errorf("extractFunctions() param count = %d, want 4", functions[0].ParamCount)
	}

	if functions[0].Name != "add" {
		t.Errorf("extractFunctions() function name = %q, want %q", functions[0].Name, "add")
	}
}

func TestExtractFunctions_VoidParams(t *testing.T) {
	lines := []string{
		"void my_function(void)",
		"{",
		"	return;",
		"}",
	}

	functions := extractFunctions(lines)
	if len(functions) != 1 {
		t.Fatalf("extractFunctions() returned %d functions, want 1", len(functions))
	}

	if functions[0].ParamCount != 0 {
		t.Errorf("extractFunctions() param count = %d, want 0 for void", functions[0].ParamCount)
	}
}

func TestGetProgressBar(t *testing.T) {
	bar := getProgressBar(50)
	if !strings.Contains(bar, "[") || !strings.Contains(bar, "]") {
		t.Error("getProgressBar() should return a string with brackets")
	}

	// Test edge cases
	bar0 := getProgressBar(0)
	if bar0 == "" {
		t.Error("getProgressBar(0) should not be empty")
	}

	bar100 := getProgressBar(100)
	if bar100 == "" {
		t.Error("getProgressBar(100) should not be empty")
	}
}

// Integration test
func TestIntegration_CompleteAnalysis(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a comprehensive test file
	testFile := filepath.Join(tmpDir, "integration_test.c")
	content := `/* Header comment */
#include <stdio.h>

#define MAX_VALUE 100
#define BUFFER_SIZE 256

int calculate_sum(int a, int b)
{
	int result;

	result = a + b;
	return result;
}

int main(void)
{
	int x;
	int y;

	x = 10;
	y = 20;
	printf("%d\n", calculate_sum(x, y));
	return 0;
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with level 1
	analyzer1 := NewAnalyzer(1)
	report1, err := analyzer1.AnalyzePath(testFile)
	if err != nil {
		t.Fatalf("AnalyzePath() level 1 error = %v", err)
	}

	if report1.TotalFiles != 1 {
		t.Errorf("Level 1: total files = %d, want 1", report1.TotalFiles)
	}

	// Test with level 2
	analyzer2 := NewAnalyzer(2)
	report2, err := analyzer2.AnalyzePath(testFile)
	if err != nil {
		t.Fatalf("AnalyzePath() level 2 error = %v", err)
	}

	if report2.TotalFiles != 1 {
		t.Errorf("Level 2: total files = %d, want 1", report2.TotalFiles)
	}

	// Level 2 should potentially find more violations than level 1
	// (though not guaranteed with this specific file)
	if len(analyzer2.rules) <= len(analyzer1.rules) {
		t.Error("Level 2 analyzer should have more rules than level 1")
	}
}

// Additional tests for better coverage

func TestOutputJSON(t *testing.T) {
	report := &Report{
		Files: []FileResult{
			{
				Filename:   "test.c",
				Violations: []Violation{},
				Score:      100.0,
				LineCount:  10,
			},
		},
		TotalScore:      100.0,
		TotalFiles:      1,
		TotalLines:      10,
		TotalViolations: 0,
		CleanFiles:      1,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	outputJSON(report)

	w.Close()
	os.Stdout = oldStdout

	var output []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			output = append(output, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	result := string(output)
	if !strings.Contains(result, `"filename"`) {
		t.Error("outputJSON should contain JSON with filename field")
	}
	if !strings.Contains(result, `"total_score"`) {
		t.Error("outputJSON should contain JSON with total_score field")
	}
}

func TestPrintReport(t *testing.T) {
	report := &Report{
		Files: []FileResult{
			{
				Filename:   "test.c",
				Violations: []Violation{},
				Score:      100.0,
				LineCount:  10,
			},
		},
		TotalScore:      100.0,
		TotalFiles:      1,
		TotalLines:      10,
		TotalViolations: 0,
		CleanFiles:      1,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printReport(report, false)

	w.Close()
	os.Stdout = oldStdout

	var output []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			output = append(output, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	result := string(output)
	if !strings.Contains(result, "Gonana") {
		t.Error("printReport should contain header with Gonana")
	}
	if !strings.Contains(result, "test.c") {
		t.Error("printReport should contain filename")
	}
}

func TestPrintReportVerbose(t *testing.T) {
	report := &Report{
		Files: []FileResult{
			{
				Filename: "bad_file.c",
				Violations: []Violation{
					{
						Rule:        "C-L1",
						Message:     "Line too long",
						Line:        5,
						Severity:    "major",
						Description: "Line contains 100 characters",
					},
				},
				Score:     95.0,
				LineCount: 20,
			},
		},
		TotalScore:      95.0,
		TotalFiles:      1,
		TotalLines:      20,
		TotalViolations: 1,
		CleanFiles:      0,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printReport(report, true)

	w.Close()
	os.Stdout = oldStdout

	var output []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			output = append(output, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	result := string(output)
	if !strings.Contains(result, "C-L1") {
		t.Error("verbose printReport should contain rule code")
	}
	if !strings.Contains(result, "Line too long") {
		t.Error("verbose printReport should contain violation message")
	}
}

func TestPrintReportWithDifferentScores(t *testing.T) {
	tests := []struct {
		name       string
		score      float64
		expectWord string
	}{
		{"excellent score", 95.0, "EXCELLENT"},
		{"good score", 80.0, "TRÈS BIEN"},
		{"average score", 60.0, "CORRECT"},
		{"low score", 30.0, "ÉCHEC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &Report{
				Files:           []FileResult{},
				TotalScore:      tt.score,
				TotalFiles:      1,
				TotalLines:      10,
				TotalViolations: 0,
				CleanFiles:      1,
			}

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printReport(report, false)

			w.Close()
			os.Stdout = oldStdout

			var output []byte
			buf := make([]byte, 1024)
			for {
				n, err := r.Read(buf)
				if n > 0 {
					output = append(output, buf[:n]...)
				}
				if err != nil {
					break
				}
			}

			result := string(output)
			if !strings.Contains(result, tt.expectWord) {
				t.Errorf("printReport with score %.1f should contain %q", tt.score, tt.expectWord)
			}
		})
	}
}

func TestCalculateScore_EdgeCases(t *testing.T) {
	analyzer := NewAnalyzer(1)

	tests := []struct {
		name       string
		violations []Violation
		minScore   float64
		maxScore   float64
	}{
		{
			name:       "no violations",
			violations: []Violation{},
			minScore:   100.0,
			maxScore:   100.0,
		},
		{
			name: "only minor violations",
			violations: []Violation{
				{Severity: "minor"},
				{Severity: "minor"},
			},
			minScore: 96.0,
			maxScore: 96.0,
		},
		{
			name: "only major violations",
			violations: []Violation{
				{Severity: "major"},
				{Severity: "major"},
			},
			minScore: 90.0,
			maxScore: 90.0,
		},
		{
			name: "many violations",
			violations: []Violation{
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
				{Severity: "major"},
			},
			minScore: 0.0,
			maxScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := analyzer.calculateScore(tt.violations)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("calculateScore() = %f, want between %f and %f", score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestCheckFunctionCount_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected int
	}{
		{
			name: "exactly 3 functions",
			lines: []string{
				"int func1() {",
				"}",
				"int func2() {",
				"}",
				"int func3() {",
				"}",
			},
			expected: 0,
		},
		{
			name: "4 functions excluding main",
			lines: []string{
				"int func1() {",
				"}",
				"int func2() {",
				"}",
				"int func3() {",
				"}",
				"int func4() {",
				"}",
			},
			expected: 1,
		},
		{
			name: "main function not counted",
			lines: []string{
				"int main() {",
				"}",
				"int func1() {",
				"}",
			},
			expected: 0,
		},
		{
			name: "if statements not counted",
			lines: []string{
				"if (x) {",
				"}",
				"while (y) {",
				"}",
				"for (i = 0; i < 10; i++) {",
				"}",
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := &FileAnalysis{Lines: tt.lines}
			violations := checkFunctionCount(analysis, "test.c", 0)
			if len(violations) != tt.expected {
				t.Errorf("checkFunctionCount() found %d violations, want %d", len(violations), tt.expected)
			}
		})
	}
}

func TestExtractFunctions_ComplexCases(t *testing.T) {
	tests := []struct {
		name          string
		lines         []string
		expectedCount int
		checkFunc     func(*testing.T, []FunctionInfo)
	}{
		{
			name: "function with pointer return",
			lines: []string{
				"int *get_pointer(void)",
				"{",
				"	return NULL;",
				"}",
			},
			expectedCount: 1,
			checkFunc: func(t *testing.T, fns []FunctionInfo) {
				if fns[0].Name != "get_pointer" {
					t.Errorf("Expected function name 'get_pointer', got '%s'", fns[0].Name)
				}
			},
		},
		{
			name: "function declaration on separate line",
			lines: []string{
				"int my_func(int x)",
				"{",
				"	return x * 2;",
				"}",
			},
			expectedCount: 1,
			checkFunc: func(t *testing.T, fns []FunctionInfo) {
				if fns[0].StartLine != 1 {
					t.Errorf("Expected start line 1, got %d", fns[0].StartLine)
				}
				if fns[0].EndLine != 4 {
					t.Errorf("Expected end line 4, got %d", fns[0].EndLine)
				}
			},
		},
		{
			name: "nested braces in function",
			lines: []string{
				"void complex_func(void)",
				"{",
				"	if (x) {",
				"		while (y) {",
				"			do_something();",
				"		}",
				"	}",
				"}",
			},
			expectedCount: 1,
			checkFunc: func(t *testing.T, fns []FunctionInfo) {
				if fns[0].EndLine != 8 {
					t.Errorf("Expected end line 8, got %d", fns[0].EndLine)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			functions := extractFunctions(tt.lines)
			if len(functions) != tt.expectedCount {
				t.Errorf("extractFunctions() returned %d functions, want %d", len(functions), tt.expectedCount)
			}
			if tt.expectedCount > 0 && tt.checkFunc != nil {
				tt.checkFunc(t, functions)
			}
		})
	}
}

func TestCollectFiles_NonCFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create various file types
	os.WriteFile(filepath.Join(tmpDir, "test.c"), []byte("int x;"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test.h"), []byte("int x;"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test.cpp"), []byte("int x;"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("text"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("readme"), 0644)

	analyzer := NewAnalyzer(1)
	files, err := analyzer.collectFiles(tmpDir)
	if err != nil {
		t.Fatalf("collectFiles() error = %v", err)
	}

	// Should only collect .c and .h files
	if len(files) != 2 {
		t.Errorf("collectFiles() found %d files, want 2 (.c and .h only)", len(files))
	}
}

func TestAnalyzeFile_ReadError(t *testing.T) {
	analyzer := NewAnalyzer(1)

	// Try to analyze non-existent file
	_, err := analyzer.analyzeFile("/nonexistent/path/file.c")
	if err == nil {
		t.Error("analyzeFile() with non-existent file should return error")
	}
}

func TestAnalyzePath_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	analyzer := NewAnalyzer(1)
	report, err := analyzer.AnalyzePath(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzePath() error = %v", err)
	}

	if report.TotalFiles != 0 {
		t.Errorf("AnalyzePath() on empty directory should have 0 files, got %d", report.TotalFiles)
	}

	if report.TotalScore != 0.0 {
		t.Errorf("AnalyzePath() on empty directory should have 0.0 score, got %f", report.TotalScore)
	}
}

func TestAnalyzePath_NonCFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	os.WriteFile(testFile, []byte("not a c file"), 0644)

	analyzer := NewAnalyzer(1)
	report, err := analyzer.AnalyzePath(testFile)
	if err != nil {
		t.Fatalf("AnalyzePath() error = %v", err)
	}

	if report.TotalFiles != 0 {
		t.Errorf("AnalyzePath() on non-C file should have 0 files, got %d", report.TotalFiles)
	}
}

func TestAnalyzeFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.c")

	os.WriteFile(testFile, []byte(""), 0644)

	analyzer := NewAnalyzer(1)
	result, err := analyzer.analyzeFile(testFile)
	if err != nil {
		t.Fatalf("analyzeFile() error = %v", err)
	}

	if result.LineCount != 1 {
		t.Errorf("analyzeFile() on empty file should have 1 line (empty string split), got %d", result.LineCount)
	}
}

func TestReport_MultipleFilesScoreCalculation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with different quality
	file1 := filepath.Join(tmpDir, "good_file.c")
	os.WriteFile(file1, []byte("int x;"), 0644)

	file2 := filepath.Join(tmpDir, "bad_file.c")
	os.WriteFile(file2, []byte("    int y;"), 0644) // space indentation

	file3 := filepath.Join(tmpDir, "ugly_file.c")
	os.WriteFile(file3, []byte("    int a, b, c;"), 0644) // multiple violations

	analyzer := NewAnalyzer(1)
	report, err := analyzer.AnalyzePath(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzePath() error = %v", err)
	}

	if report.TotalFiles != 3 {
		t.Errorf("Report should have 3 files, got %d", report.TotalFiles)
	}

	// Total violations should be sum of all files
	if report.TotalViolations == 0 {
		t.Error("Report should have violations from bad files")
	}

	// Average score should be calculated
	if report.TotalScore == 0 || report.TotalScore > 100 {
		t.Errorf("Report total score %f should be between 0 and 100", report.TotalScore)
	}

	// Clean files should be counted correctly
	if report.CleanFiles > report.TotalFiles {
		t.Errorf("Clean files %d cannot exceed total files %d", report.CleanFiles, report.TotalFiles)
	}
}

func TestPrintHeader(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printHeader()

	w.Close()
	os.Stdout = oldStdout

	var output []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			output = append(output, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	result := string(output)
	if !strings.Contains(result, "Gonana") {
		t.Error("printHeader should contain 'Gonana'")
	}
	if !strings.Contains(result, "╔") || !strings.Contains(result, "╚") {
		t.Error("printHeader should contain box drawing characters")
	}
}

func TestPrintSummary(t *testing.T) {
	report := &Report{
		TotalFiles:      5,
		TotalLines:      100,
		TotalViolations: 10,
		CleanFiles:      3,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printSummary(report)

	w.Close()
	os.Stdout = oldStdout

	var output []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			output = append(output, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	result := string(output)
	if !strings.Contains(result, "5") {
		t.Error("printSummary should contain total files count")
	}
	if !strings.Contains(result, "100") {
		t.Error("printSummary should contain total lines")
	}
	if !strings.Contains(result, "10") {
		t.Error("printSummary should contain violations count")
	}
}

func TestCollectCFiles_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.c")
	os.WriteFile(testFile, []byte("int main() { return 0; }"), 0644)

	files, err := collectCFiles(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	if files[0] != testFile {
		t.Errorf("Expected %s, got %s", testFile, files[0])
	}
}

func TestCollectCFiles_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure
	os.WriteFile(filepath.Join(tmpDir, "file1.c"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.h"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.txt"), []byte(""), 0644)

	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "file4.c"), []byte(""), 0644)

	files, err := collectCFiles(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should find 3 C/H files (file1.c, file2.h, file4.c)
	if len(files) != 3 {
		t.Errorf("Expected 3 C files, got %d", len(files))
	}
}

func TestCollectCFiles_NonCFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("text"), 0644)

	files, err := collectCFiles(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 files for .txt file, got %d", len(files))
	}
}

func TestCollectCFiles_InvalidPath(t *testing.T) {
	_, err := collectCFiles("/nonexistent/path")
	if err == nil {
		t.Error("Expected error for nonexistent path")
	}
}

func TestRunFixer_NoCFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create non-C file
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("text"), 0644)

	analyzer := NewAnalyzer(1)
	fixer := NewFixer(analyzer, true)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runFixer(fixer, tmpDir, false)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	output := make([]byte, 1024)
	n, _ := r.Read(output)
	result := string(output[:n])

	if !strings.Contains(result, "No C files found") {
		t.Error("Expected 'No C files found' message")
	}
}

func TestRunFixer_WithVerbose(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.c")

	content := `
// Bad comment
int x, y;
`
	os.WriteFile(testFile, []byte(content), 0644)

	analyzer := NewAnalyzer(1)
	fixer := NewFixer(analyzer, true)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runFixer(fixer, tmpDir, true)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	output := make([]byte, 4096)
	n, _ := r.Read(output)
	result := string(output[:n])

	if !strings.Contains(result, "test.c") {
		t.Error("Expected file name in verbose output")
	}
	if !strings.Contains(result, "Would fix") {
		t.Error("Expected 'Would fix' in dry-run verbose output")
	}
}

func TestRunFixer_ActualFix(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.c")

	content := `
// Bad comment
int x, y;
`
	os.WriteFile(testFile, []byte(content), 0644)

	analyzer := NewAnalyzer(1)
	fixer := NewFixer(analyzer, false)

	err := runFixer(fixer, tmpDir, false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify file was modified
	modified, _ := os.ReadFile(testFile)
	modifiedStr := string(modified)

	if strings.Contains(modifiedStr, "//") {
		t.Error("Expected // comments to be converted")
	}
	if strings.Contains(modifiedStr, "int x, y;") {
		t.Error("Expected multiple declarations to be split")
	}
}

func TestRunFixer_ErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.c")

	// This should log error but continue
	content := `
int x, y;
`
	os.WriteFile(testFile, []byte(content), 0644)
	os.Chmod(testFile, 0444)

	analyzer := NewAnalyzer(1)
	fixer := NewFixer(analyzer, false)

	// This should log error but continue
	runFixer(fixer, tmpDir, false)

	// Restore permissions
	os.Chmod(testFile, 0644)

	// Just verify it doesn't panic
}

func TestAnalyzeFile_UnreadableFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.c")
	os.WriteFile(testFile, []byte("int x;"), 0644)

	// Make file unreadable
	os.Chmod(testFile, 0000)
	defer os.Chmod(testFile, 0644)

	analyzer := NewAnalyzer(1)
	result, err := analyzer.analyzeFile(testFile)

	// Should return error on read error
	if err == nil {
		t.Error("Expected error for unreadable file")
	}
	if result != nil {
		t.Error("Expected nil result for unreadable file")
	}
}

func TestCollectFiles_ErrorHandling(t *testing.T) {
	analyzer := NewAnalyzer(1)

	// Test with invalid path
	files, err := analyzer.collectFiles("/nonexistent/path/that/does/not/exist")

	// Should return error for nonexistent path
	if err == nil {
		t.Error("Expected error for nonexistent path")
	}
	if files != nil {
		t.Error("Expected nil files for error case")
	}
}

func TestPrintFileResults_LongFilename(t *testing.T) {
	report := &Report{
		Files: []FileResult{
			{
				Filename:   strings.Repeat("very_long_filename_", 10) + ".c",
				Score:      85.5,
				LineCount:  100,
				Violations: []Violation{{Rule: "C-L1", Line: 1, Message: "Test"}},
			},
		},
		TotalFiles: 1,
		TotalLines: 100,
	}

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printFileResults(report, false)

	w.Close()
	os.Stdout = oldStdout

	output := make([]byte, 4096)
	n, _ := r.Read(output)

	// Just verify it doesn't panic with long filenames
	if n == 0 {
		t.Error("Expected some output")
	}
}

func TestCalculateScore_AllViolationTypes(t *testing.T) {
	analyzer := NewAnalyzer(2)

	violations := []Violation{
		{Rule: "C-L1", Line: 1},
		{Rule: "C-L2", Line: 2},
		{Rule: "C-L3", Line: 3},
		{Rule: "C-O1", Line: 0},
		{Rule: "C-F1", Line: 5},
		{Rule: "C-C1", Line: 10},
		{Rule: "C-G1", Line: 11},
	}

	score := analyzer.calculateScore(violations)

	if score < 0 || score > 100 {
		t.Errorf("Score should be between 0 and 100, got %.1f", score)
	}

	// With multiple violations, score should be significantly reduced
	if score > 90 {
		t.Errorf("Expected lower score with multiple violations, got %.1f", score)
	}
}

func TestFixEmptyLines_EdgeCases(t *testing.T) {
	fixer := NewFixer(nil, true)

	// Test empty input
	result := &FixResult{Fixes: make([]Fix, 0)}
	fixed := fixer.fixEmptyLines([]string{}, result)
	if len(fixed) != 0 {
		t.Error("Expected empty output for empty input")
	}

	// Test single empty line
	result = &FixResult{Fixes: make([]Fix, 0)}
	fixed = fixer.fixEmptyLines([]string{""}, result)
	if len(fixed) != 0 {
		t.Error("Expected empty output for single empty line")
	}

	// Test all empty lines
	result = &FixResult{Fixes: make([]Fix, 0)}
	fixed = fixer.fixEmptyLines([]string{"", "", ""}, result)
	if len(fixed) != 0 {
		t.Error("Expected empty output for all empty lines")
	}
}

func TestFixIndentation_ComplexCases(t *testing.T) {
	fixer := NewFixer(nil, true)

	// Test 2 spaces (less than 4)
	result := &FixResult{Fixes: make([]Fix, 0)}
	fixed := fixer.fixIndentation([]string{"  int x;"}, result)
	if fixed[0] != "  int x;" {
		t.Error("Should keep spaces less than 4")
	}

	// Test 6 spaces (1 tab + 2 spaces)
	result = &FixResult{Fixes: make([]Fix, 0)}
	fixed = fixer.fixIndentation([]string{"      int x;"}, result)
	if fixed[0] != "\t  int x;" {
		t.Errorf("Expected tab + 2 spaces, got %q", fixed[0])
	}
}

func TestFixMultipleVariableDeclarations_EdgeCases(t *testing.T) {
	fixer := NewFixer(nil, true)

	// Test with pointers
	result := &FixResult{Fixes: make([]Fix, 0)}
	fixed := fixer.fixMultipleVariableDeclarations([]string{"int *x, *y;"}, result)
	// Should split even with pointers
	if len(fixed) < 2 && len(result.Fixes) > 0 {
		t.Error("Expected pointers to be split")
	}

	// Test with const
	result = &FixResult{Fixes: make([]Fix, 0)}
	fixed = fixer.fixMultipleVariableDeclarations([]string{"const int x, y;"}, result)
	// May or may not match depending on regex
	if len(fixed) == 0 {
		t.Error("Expected some output")
	}
}

func TestFixCommentFormat_EdgeCases(t *testing.T) {
	fixer := NewFixer(nil, true)

	// Test with multiple // on same line
	result := &FixResult{Fixes: make([]Fix, 0)}
	fixed := fixer.fixCommentFormat([]string{"int x; // comment // more"}, result)
	if !strings.Contains(fixed[0], "/*") {
		t.Error("Expected // to be converted")
	}

	// Test with only //
	result = &FixResult{Fixes: make([]Fix, 0)}
	fixed = fixer.fixCommentFormat([]string{"//"}, result)
	if len(fixed) == 0 {
		t.Error("Expected some output")
	}
}

func TestToSnakeCase_SpecialCases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"a", "a"},
		{"A", "a"},
		{"AB", "a_b"},
		{"ABC", "a_b_c"},
		{"Test123", "test123"},
		{"test_snake", "test_snake"},
	}

	for _, tt := range tests {
		result := toSnakeCase(tt.input)
		if result != tt.expected {
			t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
