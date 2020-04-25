package pkg

import (
	"fmt"
	"go/ast"
	"strings"
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
	case *ast.ArrayType:
		if typ.Len == nil {
			argsType += "[]"
		} else {
			argsType += fmt.Sprintf("[%s]", ParseExpr(typ.Len, fileContent))
		}
		argsType += ParseExpr(typ.Elt, fileContent)
	case *ast.MapType:
		key := ParseExpr(typ.Key, fileContent)
		val := ParseExpr(typ.Value, fileContent)
		argsType += fmt.Sprintf("map[%s]%s", key, val)
	case *ast.BasicLit:
		argsType += typ.Value
	case *ast.InterfaceType:
		argsType += "interface{}"
	case *ast.FuncType:
		argsType += strings.Trim(fileContent[typ.Pos() - 1 : typ.End()], "\n")
	default:
		panic(fmt.Sprintf("unsupport expr type %T", typ))
	}

	return argsType
}
