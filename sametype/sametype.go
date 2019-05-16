package sametype

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
FactTypes: []analysis.Fact{(*SameType)(nil)},
	Run: run,

}

type SameType struct{}

func (s *SameType) AFact() {}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	checkForFact := func(n ast.Node) {
		call := n.(*ast.CallExpr)
		fn, _ := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
		if fn == nil {
			return // not a function call
		}

		var sameType SameType
		if !pass.ImportObjectFact(fn, &sameType) {
			return
		}

		typ0 := pass.TypesInfo.Types[call.Args[0]].Type
		typ1 := pass.TypesInfo.Types[call.Args[1]].Type
		if !types.Identical(typ0, typ1) {
			pass.Reportf(call.Pos(),
				"Calls to %v must have arguments of the same type; "+
						"is called with %v and %v",
				fn.Name(), typ0, typ1)
		}}
	maybeAddFact := func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return true
		}

		call := n.(*ast.CallExpr)
		fn, _ := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
		if fn == nil {
			return false // not a function call
		}

		if fn.FullName() != "annotate.SameType" {
			return false // not an annotation
		}

		var enclosingFunc *ast.FuncDecl
		for _, node := range stack {
			if v, ok := node.(*ast.FuncDecl); ok {
				enclosingFunc = v
				break
			}
		}
		if enclosingFunc == nil  {
			return false // we didn't find the enclosing call
		} else if len(enclosingFunc.Type.Params.List) != 2 {
			pass.Reportf(call.Pos(), "SameType annotation can only be added to funcs with two arguments")
		}

		obj := pass.TypesInfo.Defs[enclosingFunc.Name]
		pass.ExportObjectFact(obj, &SameType{})
		return false
	}
	inspect.WithStack([]ast.Node{(*ast.CallExpr)(nil)}, maybeAddFact)
	inspect.Preorder(
		[]ast.Node{(*ast.CallExpr)(nil)},
		checkForFact)
	return nil, nil
}
