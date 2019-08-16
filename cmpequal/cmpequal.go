package cmpequal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
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
				pass.Reportf(call.Pos(), "arg0 is a pointer to arg1")
				fixes = append(fixes, fixDereference(pass, call.Args[0]))
			} else if isPointerTo(typ1, typ0) {
				pass.Reportf(call.Pos(), "arg1 is a pointer to arg0")
				fixes = append(fixes, fixDereference(pass, call.Args[1]))
			} else {
				pass.Reportf(call.Pos(), "isPointerTo(typ0, arg1): %v, isPointerTo(arg1, arg0): %v", isPointerTo(typ0, typ1), isPointerTo(typ1, typ0))
				pass.Reportf(call.Pos(), "isPointer(typ0): %v, isPointer(typ1): %v", isPointer(typ0), isPointer(typ1))
				if t, ok := typ1.(*types.Pointer); ok {
					pass.Reportf(call.Pos(), ".elem: %s", t.Elem())
				}
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
		return types.Identical(ptr.Elem(), b)
	}
	return false
}

func isPointer(a types.Type) bool {
	_, ok := a.(*types.Pointer)
	return ok
}

func fixDereference(pass *analysis.Pass, expr ast.Expr) analysis.SuggestedFix {
	// dereference typ0
	var buf bytes.Buffer
	if err := format.Node(&buf, pass.Fset, &ast.StarExpr{X: expr}); err != nil {
		buf.WriteString(err.Error())
	}
	fix := analysis.SuggestedFix{
		Message:   "derefenence pointer",
		TextEdits: []analysis.TextEdit{{expr.Pos(), expr.End(), buf.Bytes()}},
	}
	return fix
}
