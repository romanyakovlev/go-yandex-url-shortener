package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

func initAnalyzers() []*analysis.Analyzer {
	analyzers := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		osExitAnalyzer,
	}

	for _, a := range staticcheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	checks := map[string]bool{
		"S1000": true,
	}

	for _, a := range staticcheck.Analyzers {
		if checks[a.Analyzer.Name] {
			analyzers = append(analyzers, a.Analyzer)
		}
	}
	return analyzers
}

func main() {
	analyzers := initAnalyzers()

	multichecker.Main(
		analyzers...,
	)
}
