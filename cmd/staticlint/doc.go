// Package main is an implementation of custom tool for static analysis of Go code.
//
// This tool is a wrapper of various static analysis tools such as:
//
// go/analysis/passes/nilfunc - checks for useless comparisons between functions and nil
//
// go/analysis/passes/printf - checks consistency of Printf format strings and arguments
//
// go/analysis/passes/shadow - checks for possible unintended shadowing of variables
//
// go/analysis/passes/structtag - checks struct field tags are well-formed
//
// all SA analyzers, S1000, S1001, S1002, ST1000, ST1001.
// ST1003, QF1001, QF1002, QF1003 from staticcheck package.
// SA analyzers check various misuses of the standard library.
// S1000 checks for single-case select statements.
// S1001 checks if for loop can be replaced with a call to copy.
// S1002 checks for redundant comparisons with boolean constants.
// ST1003 checks for poorly chosen identifier names.
// ST1005 checks for incorrectly formatted error strings.
// ST1013 checks for use of constants for HTTP status codes.
// QF1001 applies De Morgan's laws to boolean expressions.
// QF1002 converts untagged switch to tagged switch.
// QF1011 omits redundant type in variable declaration.
//
// timakin/bodyclose/passes/bodyclose - checks whether res.Body is correctly closed.
//
// go-critic/go-critic/checkers/analyzer - linter with a lot of checks.
//
// custom analyzer osExitCheck - checks for calls to os.Exit in main function of main package.
//
// To run this tool, you can use the following command:
//
// staticlint [flags] packages.
//
// Use staticlint ./... to run the tool on all packages in the current directory and all of its subdirectories.
package main
