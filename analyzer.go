package main

import (
	"os"
	"path/filepath"
	"strings"
)

// Analyzer analyzes C source files for style violations
type Analyzer struct {
	level int
	rules map[string]Rule
}

// NewAnalyzer creates a new analyzer with the specified verification level
func NewAnalyzer(level int) *Analyzer {
	a := &Analyzer{
		level: level,
		rules: make(map[string]Rule),
	}
	a.initRules()
	return a
}

// initRules initializes all checking rules based on verification level
func (a *Analyzer) initRules() {
	// Level 1 rules (basic)
	a.rules["C-L1"] = Rule{
		Code: "C-L1", Name: "Line Length", Description: "Line too long (80 chars max)",
		Severity: "major", Level: 1, Check: checkLineLength,
	}
	a.rules["C-L2"] = Rule{
		Code: "C-L2", Name: "Empty Lines", Description: "Forbidden empty lines",
		Severity: "minor", Level: 1, Check: checkEmptyLines,
	}
	a.rules["C-L3"] = Rule{
		Code: "C-L3", Name: "Indentation", Description: "TAB indentation only",
		Severity: "major", Level: 1, Check: checkIndentation,
	}
	a.rules["C-L4"] = Rule{
		Code: "C-L4", Name: "Variable Declaration", Description: "One variable per line",
		Severity: "major", Level: 1, Check: checkVariableDeclaration,
	}
	a.rules["C-V1"] = Rule{
		Code: "C-V1", Name: "Variable Position", Description: "Variables at function start",
		Severity: "major", Level: 1, Check: checkVariablePosition,
	}
	a.rules["C-O1"] = Rule{
		Code: "C-O1", Name: "Filename", Description: "Filename in snake_case",
		Severity: "major", Level: 1, Check: checkFilename,
	}
	a.rules["C-O2"] = Rule{
		Code: "C-O2", Name: "Function Count", Description: "Max 3 functions per file",
		Severity: "major", Level: 1, Check: checkFunctionCount,
	}
	a.rules["C-F1"] = Rule{
		Code: "C-F1", Name: "Function Name", Description: "Function name in snake_case",
		Severity: "major", Level: 1, Check: checkFunctionNames,
	}
	a.rules["C-F2"] = Rule{
		Code: "C-F2", Name: "Macro Name", Description: "Macro in SCREAMING_SNAKE_CASE",
		Severity: "major", Level: 1, Check: checkMacroNames,
	}
	a.rules["C-F3"] = Rule{
		Code: "C-F3", Name: "Function Length", Description: "Function max 25 lines",
		Severity: "major", Level: 1, Check: checkFunctionLength,
	}

	// Level 2 rules (advanced)
	if a.level >= 2 {
		a.rules["C-C1"] = Rule{
			Code: "C-C1", Name: "Comment Format", Description: "/* */ comments only",
			Severity: "minor", Level: 2, Check: checkCommentFormat,
		}
		a.rules["C-C2"] = Rule{
			Code: "C-C2", Name: "Function Comment", Description: "Function comment required",
			Severity: "minor", Level: 2, Check: checkFunctionComment,
		}
		a.rules["C-G1"] = Rule{
			Code: "C-G1", Name: "Global Variables", Description: "No non-const globals",
			Severity: "major", Level: 2, Check: checkGlobalVariables,
		}
		a.rules["C-F4"] = Rule{
			Code: "C-F4", Name: "Function Parameters", Description: "Max 4 parameters",
			Severity: "major", Level: 2, Check: checkFunctionParameters,
		}
		a.rules["C-L5"] = Rule{
			Code: "C-L5", Name: "For Loop Declaration", Description: "No declaration in for loops",
			Severity: "major", Level: 2, Check: checkForLoopDeclaration,
		}
	}
}

// AnalyzePath analyzes a file or directory and returns a report
func (a *Analyzer) AnalyzePath(path string) (*Report, error) {
	files, err := a.collectFiles(path)
	if err != nil {
		return nil, err
	}

	report := &Report{
		Files: make([]FileResult, 0, len(files)),
	}

	for _, file := range files {
		result, err := a.analyzeFile(file)
		if err != nil {
			continue
		}
		report.Files = append(report.Files, *result)
		report.TotalFiles++
		report.TotalLines += result.LineCount
		report.TotalViolations += len(result.Violations)
		if len(result.Violations) == 0 {
			report.CleanFiles++
		}
	}

	// Calculate total score
	if report.TotalFiles > 0 {
		totalScore := 0.0
		for _, file := range report.Files {
			totalScore += file.Score
		}
		report.TotalScore = totalScore / float64(report.TotalFiles)
	}

	return report, nil
}

// collectFiles gathers all C source files from the given path
func (a *Analyzer) collectFiles(path string) ([]string, error) {
	var files []string

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(p, ".c") || strings.HasSuffix(p, ".h") {
				files = append(files, p)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else if strings.HasSuffix(path, ".c") || strings.HasSuffix(path, ".h") {
		files = append(files, path)
	}

	return files, nil
}

// analyzeFile analyzes a single file and returns its result
func (a *Analyzer) analyzeFile(filename string) (*FileResult, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	analysis := &FileAnalysis{
		Filename:  filename,
		Lines:     lines,
		Functions: extractFunctions(lines),
	}

	violations := a.checkRules(analysis, filename)
	score := a.calculateScore(violations)

	return &FileResult{
		Filename:   filepath.Base(filename),
		Violations: violations,
		Score:      score,
		LineCount:  len(lines),
	}, nil
}

// checkRules runs all applicable rules against the file
func (a *Analyzer) checkRules(analysis *FileAnalysis, filename string) []Violation {
	var violations []Violation
	for _, rule := range a.rules {
		if rule.Level <= a.level {
			ruleViolations := rule.Check(analysis, filename, 0)
			violations = append(violations, ruleViolations...)
		}
	}
	return violations
}

// calculateScore computes the file score based on violations
func (a *Analyzer) calculateScore(violations []Violation) float64 {
	score := 100.0
	for _, v := range violations {
		penalty := 5.0 // major violations
		if v.Severity == "minor" {
			penalty = 2.0
		}
		score -= penalty
	}
	if score < 0 {
		score = 0
	}
	return score
}
