package domain_file

import (
	"bytes"
	"text/template"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	"github.com/spf13/viper"
)

const (
	DOMAIN_FILE_EXT = ".go"
)

//DomainFile
type DomainFile interface {

	//if true not need this domain file
	Disabled() bool

	BindInput(*viper.Viper) error

	//parse columns information for domain file object
	ParseCloumns([]Column, *ShareInfo)

	//applies a parsed template to the domain file object
	//return Parsed content
	Execute() string

	//save the domain file content path
	GetSavePath() string

	GetName() string

	//
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

//Share information for all domain files
//avoid import cycle not allowed
type ShareInfo struct {
	//Camel Form
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

type InjectInfo struct {
	Fields pkg.Fields

	Imports pkg.Imports

	InfraSetArgs []string

	Provides Provides
}

func NewInjectInfo() *InjectInfo {
	injectInfo := &InjectInfo{}

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
	if ps.Len() < 0 {
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
