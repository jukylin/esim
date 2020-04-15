package pkg

import (
	"go/ast"
	"testing"
	"go/token"
	"go/parser"
	"github.com/stretchr/testify/assert"
)

func TestVars_String(t *testing.T) {
	var src = `package main

var a int

var (
	b string

	//dependency injection
	c = wire.NewSet(
		wire.Struct(nil, "*"),
	)

	d interface{}

	e, f, g string = "e", "f", "g"
)

var h = 5

var i, j, k string = "i", "j", "k"

var l []int

var m func(string)

var n *int
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	assert.Nil(t, err)

	vars := Vars{}
	for _, decl := range f.Decls {
		if GenDecl, ok := decl.(*ast.GenDecl); ok {
			vars.ParseFromAst(GenDecl, src)
		}
	}

	assert.Equal(t, 10, vars.Len())
}