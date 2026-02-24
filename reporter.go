package main

import (
	"fmt"
	"sort"
	"strings"
)

// printReport displays a formatted analysis report to the console
func printReport(report *Report, verbose bool) {
	printHeader()
	printSummary(report)
	printFileResults(report, verbose)
	printFinalScore(report)
}

// printHeader displays the report header
func printHeader() {
	fmt.Println(ColorBold + "╔══════════════════════════════════════════════════════════════════════════════╗" + ColorReset)
	fmt.Println(ColorBold + "║                           Gonana - RAPPORT D'ANALYSE                         ║" + ColorReset)
	fmt.Println(ColorBold + "╚══════════════════════════════════════════════════════════════════════════════╝" + ColorReset)
	fmt.Println()
}

// printSummary displays the summary statistics
func printSummary(report *Report) {
	fmt.Printf("📊 %sRÉSUMÉ GLOBAL%s\n", ColorBold, ColorReset)
	fmt.Printf("   • Fichiers analysés: %d\n", report.TotalFiles)
	fmt.Printf("   • Lignes de code: %d\n", report.TotalLines)
	fmt.Printf("   • Violations totales: %d\n", report.TotalViolations)
	fmt.Printf("   • Fichiers propres: %d/%d\n", report.CleanFiles, report.TotalFiles)

	cleanPercent := 0.0
	if report.TotalFiles > 0 {
		cleanPercent = float64(report.CleanFiles) / float64(report.TotalFiles) * 100
	}
	fmt.Printf("   • Propreté: %.1f%% %s\n", cleanPercent, getProgressBar(cleanPercent))
	fmt.Println()
}

// printFileResults displays individual file results
func printFileResults(report *Report, verbose bool) {
	// Sort files by score (descending)
	sort.Slice(report.Files, func(i, j int) bool {
		return report.Files[i].Score > report.Files[j].Score
	})

	// Print file results
	for _, file := range report.Files {
		if len(file.Violations) == 0 {
			fmt.Printf("%s✅ %s%s (%.1f%% - %d lignes)\n",
				ColorGreen, file.Filename, ColorReset, file.Score, file.LineCount)
		} else {
			fmt.Printf("%s❌ %s%s (%.1f%% - %d lignes - %d violations)\n",
				ColorRed, file.Filename, ColorReset, file.Score, file.LineCount, len(file.Violations))
		}

		if verbose && len(file.Violations) > 0 {
			printViolations(file.Violations)
		}
	}

	fmt.Println()
}

// printViolations displays detailed violation information
func printViolations(violations []Violation) {
	for _, v := range violations {
		severity := ColorYellow + "MINOR" + ColorReset
		if v.Severity == "major" {
			severity = ColorRed + "MAJOR" + ColorReset
		}
		fmt.Printf("    [%s] Line %d: %s - %s\n", severity, v.Line, v.Rule, v.Message)
		if v.Description != "" {
			fmt.Printf("         %s\n", v.Description)
		}
	}
}

// printFinalScore displays the final score and message
func printFinalScore(report *Report) {
	scoreColor := ColorRed
	scoreMessage := "ÉCHEC! Beaucoup de travail nécessaire."
	if report.TotalScore >= 90 {
		scoreColor = ColorGreen
		scoreMessage = "EXCELLENT! Code très propre."
	} else if report.TotalScore >= 75 {
		scoreColor = ColorYellow
		scoreMessage = "TRÈS BIEN! Quelques petits détails à corriger."
	} else if report.TotalScore >= 50 {
		scoreColor = ColorYellow
		scoreMessage = "CORRECT! Plusieurs améliorations nécessaires."
	}

	fmt.Println(ColorBold + "╔══════════════════════════════════════════════════════════════════════════════╗" + ColorReset)
	fmt.Printf("║%s                             SCORE GLOBAL: %.1f%%                              %s ║\n",
		scoreColor, report.TotalScore, ColorReset)
	fmt.Printf("║           %s%.1f%%           ║\n", getProgressBar(report.TotalScore), report.TotalScore)
	fmt.Printf("║                   %s                  ║\n", scoreMessage)
	fmt.Println(ColorBold + "╚══════════════════════════════════════════════════════════════════════════════╝" + ColorReset)
}

// getProgressBar generates a visual progress bar
func getProgressBar(percentage float64) string {
	barLength := 50
	filled := int(percentage / 100 * float64(barLength))
	empty := barLength - filled

	bar := ColorGreen + strings.Repeat("█", filled) + ColorReset + strings.Repeat("░", empty)
	return "[" + bar + "]"
}
