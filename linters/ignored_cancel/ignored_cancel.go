package ignoredcancel

import (
	"errors"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var IgnoredCancelAnalyzer = &analysis.Analyzer{
	Name: "ignoredcancel", // name of our analyzer
	Doc:  "linter for detecting ignored cancel function returned from context.CancelFunc",
	Run: func(p *analysis.Pass) (interface{}, error) {
		i, ok := p.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		if !ok {
			return nil, errors.New("analyzer is not of type *inspector.Inspector")
		}

		filter := []ast.Node{(*ast.AssignStmt)(nil)}
		i.Preorder(filter, func(n ast.Node) {
			foundIgnoredCancel := false // flag

			// confirm node is indeed an AssignStmt
			node, ok := n.(*ast.AssignStmt)
			if !ok {
				return
			}

			// DFS on the node's children
			ast.Inspect(node, func(n ast.Node) bool {
				// len(RHS) can only be 1 if it's a multi-return function
				// ignore all other cases
				if len(node.Rhs) > 1 {
					return false
				}

				// assert that the RHS is a function call expression
				e, ok := node.Rhs[0].(*ast.CallExpr)
				if !ok {
					return false
				}

				// assert that the function call is a selector expression
				fExpr, ok := e.Fun.(*ast.SelectorExpr)
				if !ok {
					return false
				}

				// assert that the expression in selector is an identifier
				// because it is an import of "context"
				sExpr, ok := fExpr.X.(*ast.Ident)
				if !ok {
					return false
				}

				// if the function signature matches
				if sExpr.Name != "context" || fExpr.Sel.Name != "WithCancel" {
					return false
				}

				// if lhs has more or less variables, something is very wrong
				if len(node.Lhs) != 2 {
					return false
				}

				// assert that the lhs is just an identifier
				lExpr, ok := node.Lhs[1].(*ast.Ident)
				if !ok {
					return false
				}

				if lExpr.Name == "_" {
					foundIgnoredCancel = true
					return false
				}

				return true
			})
			if foundIgnoredCancel {
				p.Reportf(n.Pos(), "found ignored cancelFunc on context.WithCancel")
			}
		})

		return nil, nil
	}, // the logic for our analyzer
	Requires: []*analysis.Analyzer{inspect.Analyzer}, // declare analyzers that ours is dependent on
}
