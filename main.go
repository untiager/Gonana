package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	// Parse command-line flags
	pathFlag := flag.String("path", "", "Path to file or directory to analyze")
	verboseFlag := flag.Bool("verbose", false, "Verbose output")
	jsonFlag := flag.Bool("json", false, "JSON output format")
	silentFlag := flag.Bool("silent", false, "Silent mode (exit code only)")
	levelFlag := flag.Int("level", 1, "Verification level (1=basic, 2=advanced)")
	flag.Parse()

	// Get path from flag or argument
	path := *pathFlag
	if path == "" && len(flag.Args()) > 0 {
		path = flag.Args()[0]
	}

	if path == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <file_or_directory>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Run analysis
	analyzer := NewAnalyzer(*levelFlag)
	report, err := analyzer.AnalyzePath(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle silent mode
	if *silentFlag {
		if report.TotalViolations > 0 {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Output results
	if *jsonFlag {
		outputJSON(report)
	} else {
		printReport(report, *verboseFlag)
	}

	// Exit with error if violations found
	if report.TotalViolations > 0 {
		os.Exit(1)
	}
}

// outputJSON prints the report in JSON format
func outputJSON(report *Report) {
	output, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(output))
}
