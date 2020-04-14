package pkg

import (
	"go/ast"
)

func ParseExpr(expr ast.Expr, fileContent string) string {
	var argsType string
	switch typ := expr.(type) {
	case *ast.CallExpr:
		argsType += ParseExpr(typ.Fun, fileContent)
		argsType += fileContent[typ.Lparen - 1 : typ.Rparen + 1]

	case *ast.SelectorExpr:
		argsType += ParseExpr(typ.X, fileContent)
		argsType += "."
		argsType += ParseExpr(typ.Sel, fileContent)

	case *ast.StarExpr:
		argsType += "*"
		argsType += ParseExpr(typ.X, fileContent)

	case *ast.Ident:
		argsType += typ.String()

	default:
		panic("unsupport expr type")
	}

	return argsType
}
