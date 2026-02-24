package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"epicstyle/internal/types"
)

// CheckLineLength validates that no line exceeds 80 characters
func CheckLineLength(analysis *types.FileAnalysis, filename string, lineNum int) []types.Violation {
	var violations []types.Violation
	for i, line := range analysis.Lines {
		if len(line) > 80 {
			violations = append(violations, types.Violation{
				Rule:        "C-L1",
				Message:     "Line too long",
				Line:        i + 1,
				Severity:    "major",
				Description: fmt.Sprintf("Line contains %d characters (max 80)", len(line)),
			})
		}
	}
	return violations
}

// checkEmptyLines checks for forbidden empty lines
func CheckEmptyLines(analysis *types.FileAnalysis, filename string, lineNum int) []types.Violation {
	var violations []types.Violation
	lines := analysis.Lines

	// Check first line
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		violations = append(violations, types.Violation{
			Rule:        "C-L2",
			Message:     "Empty line at beginning of file",
			Line:        1,
			Severity:    "minor",
			Description: "File should not start with empty line",
		})
	}

	// Check last line
	if len(lines) > 1 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		violations = append(violations, types.Violation{
			Rule:        "C-L2",
			Message:     "Empty line at end of file",
			Line:        len(lines),
			Severity:    "minor",
			Description: "File should not end with empty line",
		})
	}

	// Check consecutive empty lines
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" && strings.TrimSpace(lines[i-1]) == "" {
			violations = append(violations, types.Violation{
				Rule:        "C-L2",
				Message:     "Consecutive empty lines",
				Line:        i + 1,
				Severity:    "minor",
				Description: "Multiple consecutive empty lines are forbidden",
			})
		}
	}

	return violations
}

// checkIndentation validates that only TABs are used for indentation
func CheckIndentation(analysis *types.FileAnalysis, filename string, lineNum int) []types.Violation {
	var violations []types.Violation
	for i, line := range analysis.Lines {
		if len(line) > 0 && line[0] == ' ' {
			violations = append(violations, types.Violation{
				Rule:        "C-L3",
				Message:     "Space indentation",
				Line:        i + 1,
				Severity:    "major",
				Description: "Use TAB for indentation, not spaces",
			})
		}
	}
	return violations
}

// checkVariableDeclaration ensures only one variable per line
func CheckVariableDeclaration(analysis *types.FileAnalysis, filename string, lineNum int) []types.Violation {
	var violations []types.Violation
	for i, line := range analysis.Lines {
		trimmed := strings.TrimSpace(line)
		// Simple check for multiple variable declarations
		if strings.Contains(trimmed, "int ") || strings.Contains(trimmed, "char ") ||
			strings.Contains(trimmed, "float ") || strings.Contains(trimmed, "double ") {
			if strings.Count(trimmed, ",") > 0 && !strings.Contains(trimmed, "for") {
				violations = append(violations, types.Violation{
					Rule:        "C-L4",
					Message:     "Multiple variable declaration",
					Line:        i + 1,
					Severity:    "major",
					Description: "Declare only one variable per line",
				})
			}
		}
	}
	return violations
}

// checkVariablePosition validates variables are at function start
func CheckVariablePosition(analysis *types.FileAnalysis, filename string, lineNum int) []types.Violation {
	// This is a simplified check - would need proper C parsing for accuracy
	return []types.Violation{}
}

// checkFilename validates that filename is in snake_case
func CheckFilename(analysis *types.FileAnalysis, filename string, lineNum int) []types.Violation {
	var violations []types.Violation
	base := filepath.Base(filename)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	if !types.IsSnakeCase(name) {
		violations = append(violations, types.Violation{
			Rule:        "C-O1",
			Message:     "Invalid filename format",
			Line:        0,
			Severity:    "major",
			Description: "Filename must be in snake_case",
		})
	}
	return violations
}

// checkFunctionCount ensures max 3 functions per file (excluding main)
func CheckFunctionCount(analysis *types.FileAnalysis, filename string, lineNum int) []types.Violation {
	var violations []types.Violation
	funcCount := 0

	for _, line := range analysis.Lines {
		trimmed := strings.TrimSpace(line)
		// Simple function detection
		if strings.Contains(trimmed, "(") && strings.Contains(trimmed, ")") &&
			strings.Contains(trimmed, "{") && !strings.HasPrefix(trimmed, "//") &&
			!strings.HasPrefix(trimmed, "/*") && !strings.Contains(trimmed, "if") &&
			!strings.Contains(trimmed, "while") && !strings.Contains(trimmed, "for") {
			if !strings.Contains(trimmed, "main") {
				funcCount++
			}
		}
	}

	if funcCount > 3 {
		violations = append(violations, types.Violation{
			Rule:        "C-O2",
			Message:     "Too many functions",
			Line:        0,
			Severity:    "major",
			Description: fmt.Sprintf("File contains %d functions (max 3 excluding main)", funcCount),
		})
	}
	return violations
}

// checkFunctionNames validates function names are in snake_case
func CheckFunctionNames(analysis *types.FileAnalysis, filename string, lineNum int) []types.Violation {
	var violations []types.Violation
	for _, fn := range analysis.Functions {
		if !types.IsSnakeCase(fn.Name) && fn.Name != "main" {
			violations = append(violations, types.Violation{
				Rule:        "C-F1",
				Message:     "Invalid function name",
				Line:        fn.StartLine,
				Severity:    "major",
				Description: fmt.Sprintf("Function '%s' must be in snake_case", fn.Name),
			})
		}
	}
	return violations
}

// checkMacroNames validates macro names are in SCREAMING_SNAKE_CASE
func CheckMacroNames(analysis *types.FileAnalysis, filename string, lineNum int) []types.Violation {
	var violations []types.Violation
	for i, line := range analysis.Lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#define ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				macroName := parts[1]
				if !types.IsScreamingSnakeCase(macroName) {
					violations = append(violations, types.Violation{
						Rule:        "C-F2",
						Message:     "Invalid macro name",
						Line:        i + 1,
						Severity:    "major",
						Description: fmt.Sprintf("Macro '%s' must be in SCREAMING_SNAKE_CASE", macroName),
					})
				}
			}
		}
	}
	return violations
}

// checkFunctionLength validates functions don't exceed 25 lines
func CheckFunctionLength(analysis *types.FileAnalysis, filename string, lineNum int) []types.Violation {
	var violations []types.Violation
	for _, fn := range analysis.Functions {
		length := fn.EndLine - fn.StartLine + 1
		if length > 25 {
			violations = append(violations, types.Violation{
				Rule:        "C-F3",
				Message:     "Function too long",
				Line:        fn.StartLine,
				Severity:    "major",
				Description: fmt.Sprintf("Function '%s' has %d lines (max 25)", fn.Name, length),
			})
		}
	}
	return violations
}

// checkCommentFormat validates use of /* */ comments only
func CheckCommentFormat(analysis *types.FileAnalysis, filename string, lineNum int) []types.Violation {
	var violations []types.Violation
	for i, line := range analysis.Lines {
		if strings.Contains(line, "//") {
			violations = append(violations, types.Violation{
				Rule:        "C-C1",
				Message:     "Invalid comment format",
				Line:        i + 1,
				Severity:    "minor",
				Description: "Use /* */ comments only, not // comments",
			})
		}
	}
	return violations
}

// checkFunctionComment validates function comments are present
func CheckFunctionComment(analysis *types.FileAnalysis, filename string, lineNum int) []types.Violation {
	// Simplified check - would need better parsing
	return []types.Violation{}
}

// checkGlobalVariables validates no non-const globals
func CheckGlobalVariables(analysis *types.FileAnalysis, filename string, lineNum int) []types.Violation {
	// Simplified check - would need proper C parsing
	return []types.Violation{}
}

// checkFunctionParameters validates max 4 parameters per function
func CheckFunctionParameters(analysis *types.FileAnalysis, filename string, lineNum int) []types.Violation {
	var violations []types.Violation
	for _, fn := range analysis.Functions {
		if fn.ParamCount > 4 {
			violations = append(violations, types.Violation{
				Rule:        "C-F4",
				Message:     "Too many parameters",
				Line:        fn.StartLine,
				Severity:    "major",
				Description: fmt.Sprintf("Function '%s' has %d parameters (max 4)", fn.Name, fn.ParamCount),
			})
		}
	}
	return violations
}

// checkForLoopDeclaration validates no variable declarations in for loops
func CheckForLoopDeclaration(analysis *types.FileAnalysis, filename string, lineNum int) []types.Violation {
	var violations []types.Violation
	for i, line := range analysis.Lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "for") && strings.Contains(trimmed, "int ") {
			violations = append(violations, types.Violation{
				Rule:        "C-L5",
				Message:     "Variable declaration in for loop",
				Line:        i + 1,
				Severity:    "major",
				Description: "Do not declare variables in for loop initialization",
			})
		}
	}
	return violations
}
