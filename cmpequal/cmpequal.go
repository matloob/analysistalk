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
				fixes = fixDereference(pass, call.Args[0])
			} else if isPointerTo(typ1, typ0) {
				fixes = fixDereference(pass, call.Args[1])
			}
			reportWithFixes(pass, call, fixes, "\n    cmp.Equal's arguments must have the same type\n"+
				"        but it's called with \u001b[31m%s\u001b[0m and \u001b[31m%s\u001b[0m values", typeName(typ0), typeName(typ1))
		}
	}
	inspect.Preorder(
		[]ast.Node{(*ast.CallExpr)(nil)},
		inspectNode)
	return nil, nil
}

func typeName(t types.Type) string {
	switch t := t.(type) {
	case *types.Named:
		return t.Obj().Pkg().Name() + "." + t.Obj().Name()
	case *types.Pointer:
		return "*" + typeName(t.Elem())
	}
	return fmt.Sprint(t)
}

func reportWithFixes(pass *analysis.Pass, node ast.Node, fixes []analysis.SuggestedFix, format string, formatArgs ...interface{}) {
	pass.Report(analysis.Diagnostic{Pos: node.Pos(), End: node.End(), Message: fmt.Sprintf(format, formatArgs...), SuggestedFixes: fixes})
}

func isPointerTo(a, b types.Type) bool {
	if ptr, ok := a.(*types.Pointer); ok {
		return types.Identical(ptr.Elem(), b)
	}
	return false
}

func fixDereference(pass *analysis.Pass, expr ast.Expr) []analysis.SuggestedFix {
	// dereference typ0
	var buf bytes.Buffer
	format.Node(&buf, pass.Fset, &ast.StarExpr{X: expr})
	fix := analysis.SuggestedFix{
		Message:   "dereference pointer",
		TextEdits: []analysis.TextEdit{{expr.Pos(), expr.End(), buf.Bytes()}},
	}
	return []analysis.SuggestedFix{fix}
}
