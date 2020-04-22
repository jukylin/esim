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

	//Esim
	*container.Esim

	//redis
	Redis *redis.Client

	TestRepo repo.TestRepo

	TestHistoryRepo repo.TestHistoryRepo
}

var infraSet = wire.NewSet(
	wire.Struct(new(Infra), "*"),
	provideTestRepo,
	provideTestHistoryRepo,
)

func NewInfra() *Infra {
	infraOnce.Do(func() {
	})

	return onceInfra
}

// Close close the infra when app stop
func (inf *Infra) Close() {
}

func (inf *Infra) HealthCheck() []error {
	var errs []error
	return errs
}

func provideTestRepo(esim *container.Esim) repo.TestRepo {
	return repo.NewDbTestRepo(esim.Logger)
}

func provideTestHistoryRepo(esim *container.Esim) repo.TestHistoryRepo {
	return repo.NewDbTestHistoryRepo(esim.Logger)
}
