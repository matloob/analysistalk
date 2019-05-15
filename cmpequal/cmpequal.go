package cmpequal

import (
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

var Analyzer = &analysis.Analyzer {
	Name: "cmpequal",
	Doc: "Check arg types of cmp.Equal",
	Requires:
[]*analysis.Analyzer{inspect.Analyzer},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspectNode := func(n ast.Node) {
		call := n.(*ast.CallExpr)
		fn, _ := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
		if fn == nil {
			return // not a function call
		}

		if fn.FullName() != "github.com/google/go-cmp/cmp.Equal" {     // should also check cmp.Diff, etc.
			return // not a call to Equal
		}

		typ0 := pass.TypesInfo.Types[call.Args[0]].Type
		typ1 := pass.TypesInfo.Types[call.Args[1]].Type
		if !types.Identical(typ0, typ1) {
			pass.Reportf(call.Pos(),
				"cmp.Equal's arguments must have the same type; "+
						"is called with %v and %v",
				typ0, typ1)
		}	}
	inspect.Preorder(
		[]ast.Node{(*ast.CallExpr)(nil)},
		inspectNode)
	return nil, nil
}
