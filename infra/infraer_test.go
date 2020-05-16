package infra

import (
	"testing"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	file_dir "github.com/jukylin/esim/pkg/file-dir"
	domain_file "github.com/jukylin/esim/tool/db2entity/domain-file"
	"github.com/stretchr/testify/assert"
)

func TestInfraer_BuildNewInfraContent(t *testing.T) {

	expected := `package infra

import (

	//sync
	//is a test
	"sync"

	"github.com/google/wire"
	"github.com/jukylin/esim/container"
	"github.com/jukylin/esim/redis"
)

var infraOnce sync.Once
var onceInfra *Infra

type Infra struct {

	//Esim
	*container.Esim

	//redis
	Redis redis.RedisClient

	check bool

	a int
}

var infraSet = wire.NewSet(
	wire.Struct(new(Infra), "*"),
	provideA,
)

func NewInfra() *Infra {
	infraOnce.Do(func() {
	})

	return onceInfra
}

// Close close the infra when app stop
func (this *Infra) Close() {
}

func (this *Infra) HealthCheck() []error {
	var errs []error
	return errs
}
func provideA() { println("test") }
`

	infraer := NewInfraer(
		WithIfacerLogger(log.NewLogger()),
		WithIfacerWriter(file_dir.NewEsimWriter()),
		WithIfacerInfraInfo(NewInfo()),
		WithIfacerExecer(pkg.NewNullExec()),
	)

	assert.True(t, infraer.parseInfra(infraContent))

	injectInfo := domain_file.NewInjectInfo()
	injectInfo.Imports = append(injectInfo.Imports, pkg.Import{Path: "time"})
	injectInfo.Fields = append(injectInfo.Fields, pkg.Field{Field: "a int"})
	injectInfo.InfraSetArgs = append(injectInfo.InfraSetArgs, "provideA")
	injectInfo.Provides = append(injectInfo.Provides,
		domain_file.Provide{`func provideA() {println("test")}`})

	infraer.injectInfos = append(infraer.injectInfos, injectInfo)

	infraer.copyInfraInfo()

	infraer.processNewInfra()

	infraer.toStringNewInfra()

	infraer.buildNewInfraContent()

	assert.Equal(t, expected, infraer.makeCodeBeautiful(infraer.newInfraInfo.content))
}
