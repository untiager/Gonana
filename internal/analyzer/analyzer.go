package analyzer

import (
	"os"
	"path/filepath"
	"strings"

	"epicstyle/internal/rules"
	"epicstyle/internal/types"
)

// Analyzer analyzes C source files for style violations
type Analyzer struct {
	level int
	rules map[string]types.Rule
}

// NewAnalyzer creates a new analyzer with the specified verification level
func NewAnalyzer(level int) *Analyzer {
	a := &Analyzer{
		level: level,
		rules: make(map[string]types.Rule),
	}
	a.initRules()
	return a
}

// Level returns the verification level
func (a *Analyzer) Level() int {
	return a.level
}

// Rules returns the rule map
func (a *Analyzer) Rules() map[string]types.Rule {
	return a.rules
}

// initRules initializes all checking rules based on verification level
func (a *Analyzer) initRules() {
	// Level 1 rules (basic)
	a.rules["C-L1"] = types.Rule{
		Code: "C-L1", Name: "Line Length", Description: "Line too long (80 chars max)",
		Severity: "major", Level: 1, Check: rules.CheckLineLength,
	}
	a.rules["C-L2"] = types.Rule{
		Code: "C-L2", Name: "Empty Lines", Description: "Forbidden empty lines",
		Severity: "minor", Level: 1, Check: rules.CheckEmptyLines,
	}
	a.rules["C-L3"] = types.Rule{
		Code: "C-L3", Name: "Indentation", Description: "TAB indentation only",
		Severity: "major", Level: 1, Check: rules.CheckIndentation,
	}
	a.rules["C-L4"] = types.Rule{
		Code: "C-L4", Name: "Variable Declaration", Description: "One variable per line",
		Severity: "major", Level: 1, Check: rules.CheckVariableDeclaration,
	}
	a.rules["C-V1"] = types.Rule{
		Code: "C-V1", Name: "Variable Position", Description: "Variables at function start",
		Severity: "major", Level: 1, Check: rules.CheckVariablePosition,
	}
	a.rules["C-O1"] = types.Rule{
		Code: "C-O1", Name: "Filename", Description: "Filename in snake_case",
		Severity: "major", Level: 1, Check: rules.CheckFilename,
	}
	a.rules["C-O2"] = types.Rule{
		Code: "C-O2", Name: "Function Count", Description: "Max 3 functions per file",
		Severity: "major", Level: 1, Check: rules.CheckFunctionCount,
	}
	a.rules["C-F1"] = types.Rule{
		Code: "C-F1", Name: "Function Name", Description: "Function name in snake_case",
		Severity: "major", Level: 1, Check: rules.CheckFunctionNames,
	}
	a.rules["C-F2"] = types.Rule{
		Code: "C-F2", Name: "Macro Name", Description: "Macro in SCREAMING_SNAKE_CASE",
		Severity: "major", Level: 1, Check: rules.CheckMacroNames,
	}
	a.rules["C-F3"] = types.Rule{
		Code: "C-F3", Name: "Function Length", Description: "Function max 25 lines",
		Severity: "major", Level: 1, Check: rules.CheckFunctionLength,
	}

	// Level 2 rules (advanced)
	if a.level >= 2 {
		a.rules["C-C1"] = types.Rule{
			Code: "C-C1", Name: "Comment Format", Description: "/* */ comments only",
			Severity: "minor", Level: 2, Check: rules.CheckCommentFormat,
		}
		a.rules["C-C2"] = types.Rule{
			Code: "C-C2", Name: "Function Comment", Description: "Function comment required",
			Severity: "minor", Level: 2, Check: rules.CheckFunctionComment,
		}
		a.rules["C-G1"] = types.Rule{
			Code: "C-G1", Name: "Global Variables", Description: "No non-const globals",
			Severity: "major", Level: 2, Check: rules.CheckGlobalVariables,
		}
		a.rules["C-F4"] = types.Rule{
			Code: "C-F4", Name: "Function Parameters", Description: "Max 4 parameters",
			Severity: "major", Level: 2, Check: rules.CheckFunctionParameters,
		}
		a.rules["C-L5"] = types.Rule{
			Code: "C-L5", Name: "For Loop Declaration", Description: "No declaration in for loops",
			Severity: "major", Level: 2, Check: rules.CheckForLoopDeclaration,
		}
	}
}

// AnalyzePath analyzes a file or directory and returns a report
func (a *Analyzer) AnalyzePath(path string) (*types.Report, error) {
	files, err := a.CollectFiles(path)
	if err != nil {
		return nil, err
	}

	report := &types.Report{
		Files: make([]types.FileResult, 0, len(files)),
	}

	for _, file := range files {
		result, err := a.AnalyzeFile(file)
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
// CollectFiles collects all C/H files from the given path
func (a *Analyzer) CollectFiles(path string) ([]string, error) {
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

// AnalyzeFile analyzes a single file and returns its result
func (a *Analyzer) AnalyzeFile(filename string) (*types.FileResult, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	analysis := &types.FileAnalysis{
		Filename:  filename,
		Lines:     lines,
		Functions: types.ExtractFunctions(lines),
	}

	violations := a.checkRules(analysis, filename)
	score := a.CalculateScore(violations)

	return &types.FileResult{
		Filename:   filepath.Base(filename),
		Violations: violations,
		Score:      score,
		LineCount:  len(lines),
	}, nil
}

// checkRules runs all applicable rules against the file
func (a *Analyzer) checkRules(analysis *types.FileAnalysis, filename string) []types.Violation {
	var violations []types.Violation
	for _, rule := range a.rules {
		if rule.Level <= a.level {
			ruleViolations := rule.Check(analysis, filename, 0)
			violations = append(violations, ruleViolations...)
		}
	}
	return violations
}

// CalculateScore computes the file score based on violations
func (a *Analyzer) CalculateScore(violations []types.Violation) float64 {
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
