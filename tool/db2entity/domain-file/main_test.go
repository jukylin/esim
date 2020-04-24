package domain_file

import (
	"os"
	"testing"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg/templates"
)

var (

	testEntityDomainFile DomainFile

	testDaoDomainFile DomainFile

	testRepoDomainFile DomainFile
)


func TestMain(m *testing.M) {
	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))

	tpl := templates.NewTextTpl()

	testEntityDomainFile = NewEntityDomainFile(
		WithEntityDomainFileLogger(logger),
		WithEntityDomainFileTpl(tpl))


	testDaoDomainFile = NewDaoDomainFile(
		WithDaoDomainFileLogger(logger),
		WithDaoDomainFileTpl(tpl))

	testRepoDomainFile = NewRepoDomainFile(
		WithRepoDomainFileLogger(logger),
		WithRepoDomainFileTpl(tpl))

	code := m.Run()

	os.Exit(code)
}



