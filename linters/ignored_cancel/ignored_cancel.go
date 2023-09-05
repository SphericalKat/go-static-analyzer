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
			switch node := n.(type) {
			case *ast.AssignStmt:
				// DFS on the node's children
				ast.Inspect(node, func(n ast.Node) bool {
					// loop through all the RHS nodes
					for _, expr := range node.Rhs {
						switch e := expr.(type) {
						// check if any are CallExprs
						case *ast.CallExpr:
							switch fExpr := e.Fun.(type) {
							// if yes, check what kind of function was called. 
							case *ast.SelectorExpr:
								switch sExpr := fExpr.X.(type) {
								case *ast.Ident:
									if (sExpr.Name == "context" && fExpr.Sel.Name == "WithCancel") {
										for i, lhsExpr := range node.Lhs {
											switch lExpr := lhsExpr.(type) {
											case *ast.Ident:
												if i == 1 && lExpr.Name == "_" {
													foundIgnoredCancel = true
													return false
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
			if foundIgnoredCancel {
				p.Reportf(n.Pos(), "found ignored cancelFunc on context.WithCancel")
			}
		})

		return nil, nil
	}, // the logic for our analyzer
	Requires: []*analysis.Analyzer{inspect.Analyzer}, // declare analyzers that ours is dependent on
}
