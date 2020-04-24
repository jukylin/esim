package domain_file

import (
	"github.com/spf13/viper"
	"github.com/jukylin/esim/tool/db2entity"
	"github.com/jukylin/esim/log"
)


const (
	DOMAIN_File_EXT = ".go"
)

//DomainFile
type DomainFile interface {

	//if true not need the domain file
	Disabled() bool

	BindInput(*viper.Viper) error

	//parse columns Information for domain file object
	ParseCloumns([]Column, *db2entity.Db2Entity)

	//applies a parsed template to the domain file object
	//return Parsed content
	Execute() string

	//save the domain file content path
	GetSavePath() string
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