package infra

var (
	infraContent = `package infra

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
}


var infraSet = wire.NewSet(
	wire.Struct(new(Infra), "*"),
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
}`
)
