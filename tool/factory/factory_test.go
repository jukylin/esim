package factory

import (
	"go/token"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/dave/dst"
	"github.com/jukylin/esim/log"
	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const (
	testStructName = "Test"
)

var resultExpectd = `package example

import (
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
)

var (
	var1 = []string{"var1"} //nolint:unused,varcheck,deadcode
)

//nolint:unused,structcheck,maligned
type Test struct {
	b int64

	c int8

	i bool

	f float32

	a int32

	h []int

	hh []interface{}

	m map[string]interface{}

	e string

	g byte

	u [3]string

	d int16

	pkg.Fields

	pkg.Field

	n func(interface{})

	o uint

	p complex64

	q rune

	r uintptr

	logger log.Logger

	conf config.Config
}

type TestOption func(*Test)

func NewTest(options ...TestOption) *Test {
	t := testPool.Get().(*Test)

	for _, option := range options {
		option(t)
	}

	if t.h == nil {
		t.h = make([]int, 0)
	}

	if t.hh == nil {
		t.hh = make([]interface{}, 0)
	}

	if t.m == nil {
		t.m = make(map[string]interface{}, 0)
	}

	if t.u == nil {
		t.u = make([3]string, 0)
	}

	return t
}

type Test1 struct {
	a int
}
`

var extendExcept = `package example

import (
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
)

var (
	var1 = []string{"var1"} //nolint:unused,varcheck,deadcode
)

//nolint:unused,structcheck,maligned
type Test struct {
	b int64

	c int8

	i bool

	f float32

	a int32

	h []int

	m map[string]interface{}

	e string

	g byte

	u [3]string

	d int16

	pkg.Fields

	pkg.Field

	n func(interface{})

	logger log.Logger

	conf config.Config
}
`

func TestMain(m *testing.M) {
	code := m.Run()

	os.Exit(code)
}

func TestEsimFactory_NotFound(t *testing.T) {
	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))
	esimfactory := NewEsimFactory(
		WithEsimFactoryLogger(logger),
		WithEsimFactoryWriter(filedir.NewEsimWriter()),
	)

	v := viper.New()
	v.Set("sname", testStructName+"1")
	v.Set("sdir", "./example")

	assert.Panics(t, func() {
		esimfactory.Run(v)
	})
}

func TestEsimFactory_ExtendFieldAndSortField(t *testing.T) {
	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))
	esimfactory := NewEsimFactory(
		WithEsimFactoryLogger(logger),
	)

	esimfactory.structDir = "./example"
	esimfactory.StructName = testStructName
	esimfactory.withOption = true
	esimfactory.withGenLoggerOption = true
	esimfactory.withGenConfOption = true
	esimfactory.WithNew = true
	esimfactory.withStar = true
	esimfactory.withPrint = true

	esimfactory.UpStructName = templates.FirstToUpper(testStructName)
	esimfactory.ShortenStructName = templates.Shorten(testStructName)
	esimfactory.LowerStructName = strings.ToLower(testStructName)

	ps := esimfactory.loadPackages()

	found := esimfactory.findStruct(ps)
	assert.True(t, found)

	esimfactory.withSort = true
	esimfactory.sortField()

	optionDecl := esimfactory.constructOptionTypeFunc()
	esimfactory.newDecls = append(esimfactory.newDecls, optionDecl)

	funcDecl := esimfactory.constructNew()
	esimfactory.newDecls = append(esimfactory.newDecls, funcDecl)

	esimfactory.extendFields()
	esimfactory.constructDecls()
	println(esimfactory.newContext())
	assert.Equal(t, resultExpectd, esimfactory.newContext())
}

func TestEsimFactory_getNewFuncTypeReturn(t *testing.T) {
	tests := []struct {
		name          string
		structName    string
		withPool      bool
		withStar      bool
		withInterface bool
		InterName     string
		want          interface{}
	}{
		{"normal", "Test", false, false, false, "",
			dst.NewIdent("Test")},
		{"with pool", "Test", true,
			false, false, "", &dst.StarExpr{
				X: dst.NewIdent("Test")}},
		{"with star", "Test", false,
			true, false, "", &dst.StarExpr{
				X: dst.NewIdent("Test")}},
		{"with interface", "", false,
			false, true, "Test", dst.NewIdent("Test")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ef := NewEsimFactory()
			ef.withStar = tt.withStar
			ef.withPool = tt.withPool
			ef.StructName = tt.structName
			ef.withImpIface = tt.InterName
			got := ef.getNewFuncTypeReturn()
			assert.True(t, reflect.DeepEqual(got.List[0].Type, tt.want))
		})
	}
}

func TestEsimFactory_getStructInstan(t *testing.T) {
	tests := []struct {
		name       string
		structName string
		withPool   bool
		withStar   bool
		want       interface{}
	}{
		{"normal", "Test", false, false,
			&dst.CompositeLit{
				Type: dst.NewIdent("Test"),
			}},
		{"with pool", "Test", true,
			false, &dst.TypeAssertExpr{
				X: &dst.CallExpr{
					Fun: &dst.SelectorExpr{
						X:   dst.NewIdent("testPool"),
						Sel: dst.NewIdent("Get"),
					},
				},
				Type: &dst.StarExpr{
					X: dst.NewIdent("Test"),
				},
			}},
		{"with star", "Test", false,
			true, &dst.UnaryExpr{
				Op: token.AND,
				X: &dst.CompositeLit{
					Type: dst.NewIdent("Test"),
				},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ef := NewEsimFactory()
			ef.withStar = tt.withStar
			ef.withPool = tt.withPool
			ef.ShortenStructName = "t"
			ef.LowerStructName = "test"
			ef.StructName = tt.structName
			ef.UpStructName = "Test"
			got := ef.getStructInstan()

			assert.True(t, reflect.DeepEqual(got.(*dst.AssignStmt).Rhs[0], tt.want))
		})
	}
}