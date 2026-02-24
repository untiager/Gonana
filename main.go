package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Parse command-line flags
	pathFlag := flag.String("path", "", "Path to file or directory to analyze")
	verboseFlag := flag.Bool("verbose", false, "Verbose output")
	jsonFlag := flag.Bool("json", false, "JSON output format")
	silentFlag := flag.Bool("silent", false, "Silent mode (exit code only)")
	levelFlag := flag.Int("level", 1, "Verification level (1=basic, 2=advanced)")
	fixFlag := flag.Bool("fix", false, "Automatically fix violations")
	dryRunFlag := flag.Bool("dry-run", false, "Show what would be fixed without applying changes")
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

	// Handle fix mode
	if *fixFlag || *dryRunFlag {
		fixer := NewFixer(analyzer, *dryRunFlag)
		if err := runFixer(fixer, path, *verboseFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

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

// runFixer runs the fixer on the given path
func runFixer(fixer *Fixer, path string, verbose bool) error {
	// Get list of C files to fix
	files, err := collectCFiles(path)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Println("No C files found to fix")
		return nil
	}

	totalFixes := 0
	filesModified := 0

	// Process each file
	for _, file := range files {
		result, err := fixer.FixFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fixing %s: %v\n", file, err)
			continue
		}

		if len(result.Fixes) > 0 {
			totalFixes += len(result.Fixes)
			if result.ModifiedContent {
				filesModified++
			}

			// Print fixes
			if verbose || fixer.dryRun {
				fmt.Printf("\n%s%s%s\n", ColorBlue, result.Filename, ColorReset)
				for _, fix := range result.Fixes {
					mode := "Fixed"
					if fixer.dryRun {
						mode = "Would fix"
					}
					if fix.Line > 0 {
						fmt.Printf("  %s [%s] Line %d: %s\n", mode, fix.Rule, fix.Line, fix.Description)
					} else {
						fmt.Printf("  %s [%s] %s\n", mode, fix.Rule, fix.Description)
					}
				}
			}

			// Handle file rename
			if result.NewFilename != "" {
				if !fixer.dryRun {
					if err := os.Rename(file, result.NewFilename); err != nil {
						fmt.Fprintf(os.Stderr, "Error renaming %s to %s: %v\n", file, result.NewFilename, err)
					} else if verbose {
						fmt.Printf("  Renamed: %s -> %s\n", result.Filename, filepath.Base(result.NewFilename))
					}
				} else if verbose {
					fmt.Printf("  Would rename: %s -> %s\n", result.Filename, filepath.Base(result.NewFilename))
				}
			}
		}
	}

	// Print summary
	fmt.Printf("\n%sSummary:%s\n", ColorBold, ColorReset)
	fmt.Printf("  Files processed: %d\n", len(files))
	if fixer.dryRun {
		fmt.Printf("  Fixes available: %d\n", totalFixes)
		if totalFixes > 0 {
			fmt.Printf("\n%sRun with --fix to apply these changes%s\n", ColorYellow, ColorReset)
		}
	} else {
		fmt.Printf("  Files modified: %d\n", filesModified)
		fmt.Printf("  Total fixes applied: %d\n", totalFixes)
		if totalFixes > 0 {
			fmt.Printf("\n%s✓ Auto-fix complete%s\n", ColorGreen, ColorReset)
		}
	}

	return nil
}

// collectCFiles collects all C files from the given path
func collectCFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var files []string

	if info.IsDir() {
		err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			ext := filepath.Ext(p)
			if !info.IsDir() && (ext == ".c" || ext == ".h") {
				files = append(files, p)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		ext := filepath.Ext(path)
		if ext == ".c" || ext == ".h" {
			files = append(files, path)
		}
	}

	return files, nil
}
