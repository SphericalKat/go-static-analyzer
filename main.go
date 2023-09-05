package main

import (
	ignoredcancel "analyzer/linters/ignored_cancel"

	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(ignoredcancel.IgnoredCancelAnalyzer)
}