package main

import "strings"

// isSnakeCase checks if a string is in snake_case format
func isSnakeCase(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			return false
		}
		if r == '_' && (i == 0 || i == len(s)-1) {
			return false
		}
	}
	return true
}

// isScreamingSnakeCase checks if a string is in SCREAMING_SNAKE_CASE format
func isScreamingSnakeCase(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if r >= 'a' && r <= 'z' {
			return false
		}
		if r == '_' && (i == 0 || i == len(s)-1) {
			return false
		}
	}
	return true
}

// extractFunctions parses lines of C code to extract function information
func extractFunctions(lines []string) []FunctionInfo {
	var functions []FunctionInfo
	var currentFunc *FunctionInfo
	braceCount := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Simple function detection
		if strings.Contains(trimmed, "(") && strings.Contains(trimmed, ")") &&
			(strings.Contains(trimmed, "{") || (i+1 < len(lines) && strings.Contains(strings.TrimSpace(lines[i+1]), "{"))) {
			if !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "/*") &&
				!strings.Contains(trimmed, "if") && !strings.Contains(trimmed, "while") &&
				!strings.Contains(trimmed, "for") && !strings.Contains(trimmed, "switch") {

				// Extract function name
				parenPos := strings.Index(trimmed, "(")
				if parenPos > 0 {
					funcPart := trimmed[:parenPos]
					parts := strings.Fields(funcPart)
					if len(parts) > 0 {
						funcName := parts[len(parts)-1]
						if strings.Contains(funcName, "*") {
							funcName = strings.TrimLeft(funcName, "*")
						}

						// Count parameters
						paramPart := trimmed[parenPos+1:]
						closeParenPos := strings.Index(paramPart, ")")
						if closeParenPos > 0 {
							params := paramPart[:closeParenPos]
							paramCount := 0
							if strings.TrimSpace(params) != "" && strings.TrimSpace(params) != "void" {
								paramCount = strings.Count(params, ",") + 1
							}

							currentFunc = &FunctionInfo{
								Name:       funcName,
								StartLine:  i + 1,
								ParamCount: paramCount,
							}
						}
					}
				}
			}
		}

		// Count braces to find function end
		braceCount += strings.Count(line, "{")
		braceCount -= strings.Count(line, "}")

		if currentFunc != nil && braceCount == 0 && strings.Contains(line, "}") {
			currentFunc.EndLine = i + 1
			functions = append(functions, *currentFunc)
			currentFunc = nil
		}
	}

	return functions
}
