package infra

import (
	"testing"

	"github.com/dave/dst"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	filedir "github.com/jukylin/esim/pkg/file-dir"
	domain_file "github.com/jukylin/esim/tool/db2entity/domain-file"
	"github.com/stretchr/testify/assert"
)

func TestInfraer_BuildNewInfraContent(t *testing.T) {
	expected := `package infra

import (
	"esim"
	"sync"
	"test/internal/infra/repo"
	_interface "test/internal/interface"

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
	a     int
}

var infraSet = wire.NewSet(
	wire.Struct(new(Infra), "*"),
	provideMns,
	provideA)

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

func provideRepo(esim *container.Esim) _interface.Repo {
	var a int
	return repo.NewRepo(esim.Logger)
}

func provideA(a string) {}

func test(esim *container.Esim) _interface.Repo {
	return repo.NewRepo(esim.Logger)
}
`

	infraer := NewInfraer(
		WithIfacerLogger(log.NewLogger()),
		WithIfacerWriter(filedir.NewNullWrite()),
		WithIfacerExecer(pkg.NewNullExec()),
	)

	injectInfo := domain_file.NewInjectInfo()
	injectInfo.Imports = append(injectInfo.Imports, pkg.Import{Path: "time", Name: "time"})
	injectInfo.Fields = append(injectInfo.Fields, pkg.Field{Field: "a int", Name: "a", Type: "int"})
	injectInfo.InfraSetArgs = append(injectInfo.InfraSetArgs, "provideA")
	injectInfo.ProvideRepoFuns = append(injectInfo.ProvideRepoFuns,
		domain_file.ProvideRepoFunc{
			FuncName:    dst.NewIdent("test"),
			ParamName:   dst.NewIdent("esim"),
			ParamType:   &dst.Ident{Name: "Esim", Path: "github.com/jukylin/esim/container"},
			Result:      &dst.Ident{Name: "Repo", Path: "test/internal/interface"},
			BodyFunc:    &dst.Ident{Name: "NewRepo", Path: "test/internal/infra/repo"},
			BodyFuncArg: &dst.Ident{Name: "Logger", Path: "esim"},
		})

	infraer.injectInfos = append(infraer.injectInfos, injectInfo)
	assert.True(t, infraer.constructNewInfra(infraContent))
	assert.Equal(t, expected, expected)

	infraer.injectInfos = make([]*domain_file.InjectInfo, 0)
	assert.False(t, infraer.constructNewInfra(infraContent))

}

func TestInfraer_constructProvideFunc(t *testing.T) {
	type args struct {
		prfs []domain_file.ProvideRepoFunc
	}
	tests := []struct {
		name   string
		args   args
		except int
	}{
		{"empty", args{prfs: []domain_file.ProvideRepoFunc{}}, 0},
		{"construct success", args{prfs: []domain_file.ProvideRepoFunc{
			domain_file.ProvideRepoFunc{
				FuncName:    dst.NewIdent("test"),
				ParamName:   dst.NewIdent("esim"),
				ParamType:   &dst.Ident{Name: "Esim", Path: "github.com/jukylin/esim/container"},
				Result:      &dst.Ident{Name: "Repo", Path: "test/internal/interface"},
				BodyFunc:    &dst.Ident{Name: "NewRepo", Path: "test/internal/infra/repo"},
				BodyFuncArg: &dst.Ident{Name: "Logger", Path: "esim"},
			},
		}}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			injectInfo := &domain_file.InjectInfo{}
			ir := NewInfraer()
			injectInfo.ProvideRepoFuns = tt.args.prfs
			ir.injectInfos = append(ir.injectInfos, injectInfo)
			funcDecls := ir.constructProvideFunc()
			assert.Equal(t, tt.except, len(funcDecls))
		})
	}
}
