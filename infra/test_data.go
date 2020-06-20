package infra

var (
	infraContent = `package infra

import (
	"sync"
	"github.com/google/wire"
	"github.com/jukylin/esim/container"
	"github.com/jukylin/esim/redis"
	_interface "test/internal/interface"
)


var infraOnce sync.Once
var onceInfra *Infra


type Infra struct {
	//Esim
	*container.Esim

	//redis
	Redis redis.RedisClient

	check bool
}


var infraSet = wire.NewSet(
	wire.Struct(new(Infra), "*"),
	provideMns,
)

func NewInfra() *Infra {
	infraOnce.Do(func() {
	})

	return onceInfra
}


// Close close the infra when app stop
func (this *Infra) Close()  {
}


func (this *Infra) HealthCheck() []error {
	var errs []error
	return errs
}

func provideRepo(esim *container.Esim) _interface.Repo {
	var a int
	return repo.NewRepo(esim.Logger)
}

func provideA(a string)  {}
`
)
