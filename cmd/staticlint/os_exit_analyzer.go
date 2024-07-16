package main

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

// osExitAnalyzer checks for calls to os.Exit within the main function.
var osExitAnalyzer = &analysis.Analyzer{
	Name: "osExitInMain",
	Doc:  "check for calls to os.Exit within the main function",
	Run:  runOsExitInMain,
}

func isOsExitCall(callExpr *ast.CallExpr) bool {
	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if selExpr.Sel.Name != "Exit" {
		return false
	}
	pkgIdent, ok := selExpr.X.(*ast.Ident)
	if !ok {
		return false
	}
	return pkgIdent.Name == "os"
}

func runOsExitInMain(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		var inMainFunc bool

		ast.Inspect(file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				inMainFunc = x.Name.Name == "main"
			case *ast.CallExpr:
				if inMainFunc && isOsExitCall(x) {
					message := "Found a call to os.Exit within the main function. This is prohibited."
					pass.Reportf(x.Pos(), message)
				}
			}
			return true
		})
	}
	return nil, nil
}
