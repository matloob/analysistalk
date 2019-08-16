package cmpequal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

var Analyzer = &analysis.Analyzer{
	Name:     "cmpequal",
	Doc:      "Check arg types of cmp.Equal",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspectNode := func(n ast.Node) {
		call := n.(*ast.CallExpr)
		fn, _ := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
		if fn == nil {
			return // not a function call
		}

		if fn.FullName() != "github.com/google/go-cmp/cmp.Equal" { // should also check cmp.Diff, etc.
			return // not a call to Equal
		}

		typ0 := pass.TypesInfo.Types[call.Args[0]].Type
		typ1 := pass.TypesInfo.Types[call.Args[1]].Type
		if !types.Identical(typ0, typ1) {
			var fixes []analysis.SuggestedFix
			if isPointerTo(typ0, typ1) {
				fixes = append(fixes, fixDereference(pass, call.Args[0]))
			} else if isPointerTo(typ1, typ0) {
				fixes = append(fixes, fixDereference(pass, call.Args[1]))
			}
			pass.Report(analysis.Diagnostic{
				Pos: call.Pos(), End: call.End(),
				Message: fmt.Sprintf("cmp.Equal's arguments must have the same type; "+
					"is called with %v and %v", typ0, typ1),
				SuggestedFixes: fixes,
			})
		}
	}
	inspect.Preorder(
		[]ast.Node{(*ast.CallExpr)(nil)},
		inspectNode)
	return nil, nil
}

func isPointerTo(a, b types.Type) bool {
	if ptr, ok := a.(*types.Pointer); ok {
		return types.Identical(ptr.Underlying(), b)
	}
	return false
}

func fixDereference(pass *analysis.Pass, expr ast.Expr) analysis.SuggestedFix {
	// dereference typ0
	var buf bytes.Buffer
	ast.Fprint(&buf, pass.Fset, ast.StarExpr{X: expr}, nil)
	fix := analysis.SuggestedFix{
		Message:   "derefenence pointer",
		TextEdits: []analysis.TextEdit{{expr.Pos(), expr.End(), buf.Bytes()}},
	}
	return fix
}
