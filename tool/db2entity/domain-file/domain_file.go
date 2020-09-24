package domainfile

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/dave/dst"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	"github.com/spf13/viper"
)

const (
	DomainFileExt = ".go"
)

type DomainFile interface {
	// 不需要生成这个文件
	Disabled() bool

	BindInput(*viper.Viper) error

	// 解析字段信息
	ParseCloumns(Columns, *ShareInfo)

	// 使用模板解析领域内容，返回解析后的内容
	Execute() string

	// 保存路径
	GetSavePath() string

	GetName() string

	GetInjectInfo() *InjectInfo
}

type DbConfig struct {
	Host string

	Port int

	User string

	Password string

	Database string

	Table string
}

func NewDbConfig() *DbConfig {
	return &DbConfig{}
}

func (dc *DbConfig) ParseConfig(v *viper.Viper, logger log.Logger) {
	host := v.GetString("host")
	if host == "" {
		logger.Fatalf("host is empty")
	}
	dc.Host = host

	port := v.GetInt("port")
	if port == 0 {
		logger.Fatalf("port is 0")
	}
	dc.Port = port

	user := v.GetString("user")
	if user == "" {
		logger.Fatalf("user is empty")
	}
	dc.User = user

	password := v.GetString("password")
	dc.Password = password

	database := v.GetString("database")
	if database == "" {
		logger.Fatalf("database is empty")
	}
	dc.Database = database

	table := v.GetString("table")
	if table == "" {
		logger.Fatalf("table is empty")
	}
	dc.Table = table
}

// 和其他领域文件共享信息，避免"import cycle not allowed".
type ShareInfo struct {
	// Camel Form
	CamelStruct string

	DbConf *DbConfig

	WithEntityTarget string

	WithDaoTarget string

	WithRepoTarget string
}

func NewShareInfo() *ShareInfo {
	return &ShareInfo{}
}

func (shareInfo *ShareInfo) ParseInfo(obj interface{}) {
	switch data := obj.(type) {
	case *entityDomainFile:
		shareInfo.WithEntityTarget = data.withEntityTarget
	case *daoDomainFile:
		shareInfo.WithDaoTarget = data.withDaoTarget
	case *repoDomainFile:
		shareInfo.WithRepoTarget = data.withRepoTarget
	}
}

type ProvideRepoFunc struct {
	FuncName *dst.Ident

	ParamName *dst.Ident

	ParamType *dst.Ident

	Result *dst.Ident

	BodyFunc *dst.Ident

	BodyFuncArg *dst.Ident
}

func NewProvideRepoFunc(entityName, path string) ProvideRepoFunc {
	provideRepoFunc := ProvideRepoFunc{}
	repoName := fmt.Sprintf("%sRepo", entityName)

	provideRepoFunc.FuncName = &dst.Ident{
		Name: fmt.Sprintf("provide%s", repoName)}
	provideRepoFunc.ParamName = dst.NewIdent("esim")
	provideRepoFunc.ParamType = &dst.Ident{
		Name: "Esim", Path: "github.com/jukylin/esim/container"}
	provideRepoFunc.Result = &dst.Ident{
		Name: repoName,
		Path: path}
	provideRepoFunc.BodyFunc = &dst.Ident{
		Name: fmt.Sprintf("NewDb%s", repoName),
		Path: path}
	provideRepoFunc.BodyFuncArg = &dst.Ident{Name: "esim.Logger"}

	return provideRepoFunc
}

type InjectInfo struct {
	Fields pkg.Fields

	Imports pkg.Imports

	InfraSetArgs []string

	// Provides Provides

	ProvideRepoFuns []ProvideRepoFunc
}

func NewInjectInfo() *InjectInfo {
	injectInfo := &InjectInfo{}

	// injectInfo.Provides = make(Provides, 0)
	injectInfo.Imports = make(pkg.Imports, 0)
	injectInfo.InfraSetArgs = make([]string, 0)
	injectInfo.ProvideRepoFuns = make([]ProvideRepoFunc, 0)

	return injectInfo
}

var provideTpl = `{{ range .Provides}}
{{.Content}}
{{end}}`

type Provide struct {
	Content string
}

type Provides []Provide

func (ps Provides) Len() int {
	return len(ps)
}

func (ps Provides) String() string {
	if ps.Len() == 0 {
		return ""
	}

	tmpl, err := template.New("provide_template").Parse(provideTpl)
	if err != nil {
		panic(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct{ Provides }{ps})
	if err != nil {
		panic(err.Error())
	}

	return buf.String()
}
