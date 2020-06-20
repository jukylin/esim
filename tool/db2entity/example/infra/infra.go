package infra

import (
	"sync"

	"github.com/google/wire"
	"github.com/jukylin/esim/container"
	"github.com/jukylin/esim/redis"
	"github.com/jukylin/esim/tool/db2entity/example/repo"
)

var infraOnce sync.Once
var onceInfra *Infra

type Infra struct {
	*container.Esim

	Redis *redis.Client

	TestHistoryRepo repo.TestHistoryRepo
}

var infraSet = wire.NewSet(
	wire.Struct(new(Infra), "*"),
	provideTestHistoryRepo)

func NewInfra() *Infra {
	infraOnce.Do(func() {
	})

	return onceInfra
}

func (inf *Infra) Close() {
}

func (inf *Infra) HealthCheck() []error {
	var errs []error
	return errs
}

func provideTestHistoryRepo(esim *container.Esim) repo.TestHistoryRepo {
	return repo.NewDbTestHistoryRepo(esim.Logger)
}
