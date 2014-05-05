package check

import (
	"fmt"
	"go/ast"
	"go/token"

	goloader "code.google.com/p/go.tools/go/loader"
	gotypes "code.google.com/p/go.tools/go/types"
)

type errorAssignmentsFinder struct{}

func NewErrorAssignmentsFinder() errorAssignmentsFinder {
	return errorAssignmentsFinder{}
}

func (c errorAssignmentsFinder) FindInAST(
	walker AstWalker,
	pkg *goloader.PackageInfo,
	file *ast.File,
	fset *token.FileSet,
) []Check {
	var checks []Check

	walker(func(n ast.Node) bool {
		// fmt.Printf("--> %#v\n", n)

		switch x := n.(type) {
		case *ast.AssignStmt:
			for _, c := range NewAssignStmtErrorAssignment(pkg, fset, x) {
				checks = append(checks, c)
			}

		case *ast.ReturnStmt:
			// errors cannot be swallowed in return

		case *ast.CallExpr:
			checks = append(checks, NewCallExprErrorAssignment(pkg, fset, x))

		case *ast.GoStmt:
			// todo

		case *ast.DeferStmt:
			// todo

		case *ast.GenDecl:
			// todo (e.g. e = errors.New("msg"))

		default:
			return true
		}

		return false
	})

	return checks
}

type funcLike interface {
	String() string
	Type() gotypes.Type
}

type errorAssignment struct {
	pkg  *goloader.PackageInfo
	fset *token.FileSet

	// Some assignment variables might be unused-untyped (_);
	// hence ast.Ident instead of gotypes.Var
	assignIdents []*ast.Ident

	funcObj   funcLike
	funcIdent *ast.Ident

	// Always types since coming from function signature
	funcReturnVars []*gotypes.Var
}

type noopErrorAssignment struct{}

// NewAssignStmtErrorAssignment constructs a check
// for function calls used with assignment op.
// e.g. a, b, := singleReturn(), singleReturn()
//      a, b, c := multiReturn()
func NewAssignStmtErrorAssignment(
	pkg *goloader.PackageInfo,
	fset *token.FileSet,
	stmt *ast.AssignStmt,
) []errorAssignment {
	var checks []errorAssignment
	var assignPos int

	for _, expr := range stmt.Rhs {
		if callExpr, ok := expr.(*ast.CallExpr); ok {
			funcObj, funcIdent, funcReturnVars := extractFunc(pkg, fset, callExpr)
			if funcObj == nil {
				continue
			}

			// Extract assignment idents corresponding to rhs return values
			assignExprs := stmt.Lhs[assignPos : assignPos+len(funcReturnVars)]
			assignIdents := extractAssignIdents(fset, assignExprs)

			checks = append(checks, errorAssignment{
				pkg:            pkg,
				fset:           fset,
				assignIdents:   assignIdents,
				funcObj:        funcObj,
				funcIdent:      funcIdent,
				funcReturnVars: funcReturnVars,
			})

			assignPos += len(funcReturnVars)
		} else {
			assignPos += 1
		}
	}

	return checks
}

// NewCallExprErrorAssignment constructs a check
// for function calls used without assignment op.
// e.g. singleReturn()
//      multiReturn()
func NewCallExprErrorAssignment(
	pkg *goloader.PackageInfo,
	fset *token.FileSet,
	expr *ast.CallExpr,
) Check {
	funcObj, funcIdent, funcReturnVars := extractFunc(pkg, fset, expr)
	if funcObj == nil {
		return noopErrorAssignment{}
	}

	return errorAssignment{
		pkg:            pkg,
		fset:           fset,
		assignIdents:   []*ast.Ident{},
		funcObj:        funcObj,
		funcIdent:      funcIdent,
		funcReturnVars: funcReturnVars,
	}
}

func (c errorAssignment) Check() ([]Problem, error) {
	var returnErrorVarIs []int
	var problems []Problem

	for i, var_ := range c.funcReturnVars {
		if obj, ok := var_.Type().(*gotypes.Named); ok {
			if obj.Obj().Name() == "error" {
				returnErrorVarIs = append(returnErrorVarIs, i)
			}
		}
	}

	for _, i := range returnErrorVarIs {
		if len(c.assignIdents) == 0 {
			problems = append(problems, Problem{
				Text:     "Return value of type error should be assigned and used",
				Package:  c.pkg.Pkg,
				Position: c.fset.Position(c.funcIdent.NamePos),
				Context: ProblemContext{
					"func": c.funcObj.String(),
				},
			})
		}

		if i < len(c.assignIdents) && c.assignIdents[i].Name == "_" {
			problems = append(problems, Problem{
				Text:     "Return value of type error should be used",
				Package:  c.pkg.Pkg,
				Position: c.fset.Position(c.assignIdents[i].NamePos),
				Context: ProblemContext{
					"func": c.funcObj.String(),
				},
			})
		}
	}

	if len(returnErrorVarIs) == 0 {
		// fmt.Printf("No errors returned from %s", c.funcObj.String())
	}

	return problems, nil
}

func (c noopErrorAssignment) Check() ([]Problem, error) {
	return []Problem{}, nil
}

// extractAssignIdents extracts variable idents on lhs of assignment
func extractAssignIdents(fset *token.FileSet, exprs []ast.Expr) []*ast.Ident {
	var idents []*ast.Ident

	for _, expr := range exprs {
		var ident *ast.Ident

		switch x := expr.(type) {

		case *ast.Ident: // e.g. ServeHTTP
			ident = x

		case *ast.SelectorExpr: // e.g. http.ServeHTTP
			ident = x.Sel

		case *ast.IndexExpr: // e.g. vals[0], vals[s.index()]
			switch y := x.X.(type) {
			case *ast.Ident:
				ident = y
			case *ast.SelectorExpr:
				ident = y.Sel
			default:
				panic(fmt.Sprintf("unknown x.X %#v", expr))
			}

		default:
			panic(fmt.Sprintf("unknown expr %#v", expr))
		}

		// non-declared-typed vars (_) will not have pkg.Defs/Uses
		idents = append(idents, ident)
	}

	return idents
}

// extractFunc extracts function defintion and return variables
func extractFunc(
	pkg *goloader.PackageInfo,
	fset *token.FileSet,
	expr *ast.CallExpr,
) (funcLike, *ast.Ident, []*gotypes.Var) {
	var funcIdent *ast.Ident

	switch x := expr.Fun.(type) {

	case *ast.Ident: // e.g. ServeHTTP(...)
		funcIdent = x

	case *ast.SelectorExpr: // e.g. http.ServeHTTP(...)
		funcIdent = x.Sel

	case *ast.ArrayType: // e.g. []byte(...)
		// No possibility of errors
		return nil, nil, nil

	case *ast.IndexExpr: // e.g. rw.beforeFuncs[i](...)
		switch y := x.X.(type) {
		case *ast.Ident:
			funcIdent = y
		case *ast.SelectorExpr:
			funcIdent = y.Sel
		default:
			panic(fmt.Sprintf("unknown x.X %#v", x.X))
		}

	case *ast.FuncLit: // e.g. func(){}(...)
		return nil, nil, nil // todo

	case *ast.ParenExpr: // ???
		return nil, nil, nil // todo

	case *ast.TypeAssertExpr: // e.g. X.(type)
		return nil, nil, nil // todo

	default:
		panic(fmt.Sprintf("unknown expr.Fun %#v", expr.Fun))
	}

	var funcObj funcLike

	switch x := pkg.Uses[funcIdent].(type) {
	case *gotypes.Func:
		funcObj = x
	case *gotypes.Var: // closure? method?
		funcObj = x
	case *gotypes.TypeName: // huh?
		return nil, nil, nil
	case *gotypes.Builtin:
		// Builtin funcs (e.g. append) do not have types
		return nil, nil, nil
	default:
		panic(fmt.Sprintf("unknown funcIdent %#v", pkg.Uses[funcIdent]))
	}

	var funcSig *gotypes.Signature

	switch x := funcObj.Type().(type) {
	case *gotypes.Signature:
		funcSig = x
	case *gotypes.Named:
		funcSig = x.Underlying().(*gotypes.Signature)
	case *gotypes.Slice: // funcObj is slice of callables
		funcSig = x.Elem().Underlying().(*gotypes.Signature)
	default:
		panic(fmt.Sprintf("not types.Signature %#v", funcObj.Type()))
	}

	var funcReturnVars []*gotypes.Var

	sigReturnVars := funcSig.Results()
	for i := 0; i < sigReturnVars.Len(); i++ {
		funcReturnVars = append(funcReturnVars, sigReturnVars.At(i))
	}

	return funcObj, funcIdent, funcReturnVars
}
