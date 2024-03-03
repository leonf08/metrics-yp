package main

import (
	"go/ast"

	"github.com/go-critic/go-critic/checkers/analyzer"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

const doc = `osExitCheckAnalyzer checks for calls to os.Exit in
main function of main package. This call is not recommended because it does not run
deferred functions, as os.Exit terminates the program immediately.`

var osExitCheck = &analysis.Analyzer{
	Name: "osExitCheck",
	Doc:  doc,
	Run:  run,
}

func main() {
	list := map[string]bool{
		"S1000":  true,
		"S1001":  true,
		"S1002":  true,
		"ST1003": true,
		"ST1005": true,
		"ST1013": true,
		"QF1001": true,
		"QF1002": true,
		"QF1011": true,
	}

	checks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		nilfunc.Analyzer,
		bodyclose.Analyzer,
		analyzer.Analyzer,
		osExitCheck,
	}

	for _, v := range staticcheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	for _, v := range simple.Analyzers {
		if list[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}

	for _, v := range stylecheck.Analyzers {
		if list[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}

	for _, v := range quickfix.Analyzers {
		if list[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}

	multichecker.Main(checks...)
}

func run(pass *analysis.Pass) (any, error) {
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			if m, ok := n.(*ast.FuncDecl); ok {
				if m.Name.Name == "main" {
					for _, b := range m.Body.List {
						if c, ok := b.(*ast.ExprStmt); ok {
							if x, ok := c.X.(*ast.CallExpr); ok {
								if s, ok := x.Fun.(*ast.SelectorExpr); ok {
									if e, ok := s.X.(*ast.Ident); ok {
										if e.Name == "os" && s.Sel.Name == "Exit" {
											pass.Reportf(x.Pos(), "os.Exit should not be called in main function")
										}
									}
								}
							}
						}
					}
				}
			}

			return true
		})
	}

	return nil, nil
}
