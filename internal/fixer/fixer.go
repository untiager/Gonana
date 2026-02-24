package fixer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"epicstyle/internal/analyzer"
	"epicstyle/internal/types"
)

// Fixer handles automatic correction of style violations
type Fixer struct {
	analyzer *analyzer.Analyzer
	dryRun   bool
}

// NewFixer creates a new fixer instance
func NewFixer(a *analyzer.Analyzer, dryRun bool) *Fixer {
	return &Fixer{
		analyzer: a,
		dryRun:   dryRun,
	}
}

// IsDryRun returns whether the fixer is in dry run mode
func (f *Fixer) IsDryRun() bool {
	return f.dryRun
}

// FixFile attempts to fix violations in a file
func (f *Fixer) FixFile(filename string) (*FixResult, error) {
	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	originalContent := string(content)
	lines := strings.Split(originalContent, "\n")

	// Track fixes applied
	result := &FixResult{
		Filename:      filepath.Base(filename),
		OriginalLines: len(lines),
		Fixes:         make([]Fix, 0),
	}

	// Apply fixes
	lines = f.fixEmptyLines(lines, result)
	lines = f.fixIndentation(lines, result)
	lines = f.fixMultipleVariableDeclarations(lines, result)
	lines = f.fixCommentFormat(lines, result)
	lines = f.fixForLoopDeclarations(lines, result)

	// Join lines back
	fixedContent := strings.Join(lines, "\n")

	// Check if filename needs fixing
	if f.shouldFixFilename(filename) {
		newName := f.fixFilename(filename)
		result.Fixes = append(result.Fixes, Fix{
			Rule:        "C-O1",
			Description: fmt.Sprintf("Rename file to %s", filepath.Base(newName)),
			Line:        0,
		})
		result.NewFilename = newName
	}

	// Only write if not dry run and content changed
	if !f.dryRun && fixedContent != originalContent {
		if err := os.WriteFile(filename, []byte(fixedContent), 0644); err != nil {
			return nil, err
		}
		result.ModifiedContent = true
	}

	result.FixedLines = len(strings.Split(fixedContent, "\n"))

	return result, nil
}

// fixEmptyLines removes forbidden empty lines (C-L2)
func (f *Fixer) fixEmptyLines(lines []string, result *FixResult) []string {
	if len(lines) == 0 {
		return lines
	}

	fixed := make([]string, 0, len(lines))

	// Remove leading empty lines
	startIdx := 0
	for startIdx < len(lines) && strings.TrimSpace(lines[startIdx]) == "" {
		result.Fixes = append(result.Fixes, Fix{
			Rule:        "C-L2",
			Description: "Removed empty line at beginning of file",
			Line:        startIdx + 1,
		})
		startIdx++
	}

	// Process middle content - remove consecutive empty lines
	prevEmpty := false
	for i := startIdx; i < len(lines); i++ {
		isEmpty := strings.TrimSpace(lines[i]) == ""

		if isEmpty && prevEmpty {
			// Skip consecutive empty line
			result.Fixes = append(result.Fixes, Fix{
				Rule:        "C-L2",
				Description: "Removed consecutive empty line",
				Line:        i + 1,
			})
			continue
		}

		fixed = append(fixed, lines[i])
		prevEmpty = isEmpty
	}

	// Remove trailing empty lines
	for len(fixed) > 0 && strings.TrimSpace(fixed[len(fixed)-1]) == "" {
		result.Fixes = append(result.Fixes, Fix{
			Rule:        "C-L2",
			Description: "Removed empty line at end of file",
			Line:        len(fixed),
		})
		fixed = fixed[:len(fixed)-1]
	}

	return fixed
}

// fixIndentation replaces leading spaces with tabs (C-L3)
func (f *Fixer) fixIndentation(lines []string, result *FixResult) []string {
	fixed := make([]string, len(lines))

	for i, line := range lines {
		if len(line) > 0 && line[0] == ' ' {
			// Count leading spaces
			spaceCount := 0
			for _, r := range line {
				if r == ' ' {
					spaceCount++
				} else {
					break
				}
			}

			// Replace with tabs (assuming 4 spaces = 1 tab)
			tabCount := spaceCount / 4
			remainder := spaceCount % 4

			fixed[i] = strings.Repeat("\t", tabCount) + strings.Repeat(" ", remainder) + line[spaceCount:]

			result.Fixes = append(result.Fixes, Fix{
				Rule:        "C-L3",
				Description: fmt.Sprintf("Replaced %d spaces with %d tabs", spaceCount, tabCount),
				Line:        i + 1,
			})
		} else {
			fixed[i] = line
		}
	}

	return fixed
}

// fixMultipleVariableDeclarations splits multiple declarations (C-L4)
func (f *Fixer) fixMultipleVariableDeclarations(lines []string, result *FixResult) []string {
	fixed := make([]string, 0, len(lines))

	varDeclRegex := regexp.MustCompile(`^\s*(int|char|float|double|long|short|unsigned)\s+([a-zA-Z_][a-zA-Z0-9_]*\s*,\s*)+([a-zA-Z_][a-zA-Z0-9_]*)\s*;`)

	for i, line := range lines {
		// Skip lines in for loops
		if strings.Contains(line, "for") {
			fixed = append(fixed, line)
			continue
		}

		if varDeclRegex.MatchString(line) {
			// Extract type and variables
			trimmed := strings.TrimSpace(line)
			parts := strings.Fields(trimmed)

			if len(parts) >= 2 {
				varType := parts[0]
				// Get the rest and remove semicolon
				varsStr := strings.TrimSuffix(strings.Join(parts[1:], " "), ";")
				vars := strings.Split(varsStr, ",")

				// Get indentation
				indent := ""
				for _, r := range line {
					if r == '\t' || r == ' ' {
						indent += string(r)
					} else {
						break
					}
				}

				// Create separate declarations
				for _, v := range vars {
					v = strings.TrimSpace(v)
					fixed = append(fixed, indent+varType+" "+v+";")
				}

				result.Fixes = append(result.Fixes, Fix{
					Rule:        "C-L4",
					Description: fmt.Sprintf("Split multiple variable declarations into %d lines", len(vars)),
					Line:        i + 1,
				})
				continue
			}
		}

		fixed = append(fixed, line)
	}

	return fixed
}

// fixCommentFormat converts // comments to /* */ (C-C1)
func (f *Fixer) fixCommentFormat(lines []string, result *FixResult) []string {
	fixed := make([]string, len(lines))

	for i, line := range lines {
		if strings.Contains(line, "//") {
			// Find the // comment
			idx := strings.Index(line, "//")
			before := line[:idx]
			comment := strings.TrimSpace(line[idx+2:])

			if comment == "" {
				fixed[i] = strings.TrimRight(before, " \t")
			} else {
				fixed[i] = before + "/* " + comment + " */"
			}

			result.Fixes = append(result.Fixes, Fix{
				Rule:        "C-C1",
				Description: "Converted // comment to /* */",
				Line:        i + 1,
			})
		} else {
			fixed[i] = line
		}
	}

	return fixed
}

// fixForLoopDeclarations extracts variable declarations from for loops (C-L5)
func (f *Fixer) fixForLoopDeclarations(lines []string, result *FixResult) []string {
	fixed := make([]string, 0, len(lines)*2)

	forDeclRegex := regexp.MustCompile(`^\s*for\s*\(\s*(int|char|float|double)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*([^;]+);(.*)$`)

	for i, line := range lines {
		matches := forDeclRegex.FindStringSubmatch(line)
		if len(matches) >= 5 {
			// Extract indentation
			indent := ""
			for _, r := range line {
				if r == '\t' || r == ' ' {
					indent += string(r)
				} else {
					break
				}
			}

			varType := matches[1]
			varName := matches[2]
			initValue := strings.TrimSpace(matches[3])
			rest := matches[4]

			// Add variable declaration
			fixed = append(fixed, indent+varType+" "+varName+";")
			fixed = append(fixed, "")

			// Add modified for loop
			forLoop := indent + "for (" + varName + " = " + initValue + ";" + rest
			fixed = append(fixed, forLoop)

			result.Fixes = append(result.Fixes, Fix{
				Rule:        "C-L5",
				Description: "Extracted variable declaration from for loop",
				Line:        i + 1,
			})
			continue
		}

		fixed = append(fixed, line)
	}

	return fixed
}

// shouldFixFilename checks if filename needs fixing (C-O1)
func (f *Fixer) shouldFixFilename(filename string) bool {
	base := filepath.Base(filename)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	return !types.IsSnakeCase(name)
}

// fixFilename converts filename to snake_case (C-O1)
func (f *Fixer) fixFilename(filename string) string {
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// Convert to snake_case
	snakeName := types.ToSnakeCase(name)

	return filepath.Join(dir, snakeName+ext)
}

// Fix represents a single fix applied
type Fix struct {
	Rule        string
	Description string
	Line        int
}

// FixResult contains the results of fixing a file
type FixResult struct {
	Filename        string
	OriginalLines   int
	FixedLines      int
	Fixes           []Fix
	ModifiedContent bool
	NewFilename     string
}
