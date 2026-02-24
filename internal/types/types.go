package types

// Violation represents a single coding style violation
type Violation struct {
	Rule        string `json:"rule"`
	Message     string `json:"message"`
	Line        int    `json:"line"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

// FileResult contains the analysis results for a single file
type FileResult struct {
	Filename   string      `json:"filename"`
	Violations []Violation `json:"violations"`
	Score      float64     `json:"score"`
	LineCount  int         `json:"line_count"`
}

// Report contains the overall analysis results
type Report struct {
	Files           []FileResult `json:"files"`
	TotalScore      float64      `json:"total_score"`
	TotalFiles      int          `json:"total_files"`
	TotalLines      int          `json:"total_lines"`
	TotalViolations int          `json:"total_violations"`
	CleanFiles      int          `json:"clean_files"`
}

// FileAnalysis contains the parsed content of a file
type FileAnalysis struct {
	Filename  string
	Lines     []string
	Functions []FunctionInfo
}

// FunctionInfo contains information about a function in the code
type FunctionInfo struct {
	Name       string
	StartLine  int
	EndLine    int
	ParamCount int
}

// Rule represents a code style rule with its checking logic
type Rule struct {
	Code        string
	Name        string
	Description string
	Severity    string
	Level       int
	Check       func(*FileAnalysis, string, int) []Violation
}
