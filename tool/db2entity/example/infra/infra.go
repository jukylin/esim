package infra

import (
	"sync"

	"github.com/google/wire"
	"github.com/jukylin/esim/container"
	"github.com/jukylin/esim/redis"
)

var infraOnce sync.Once
var onceInfra *Infra

type Infra struct {
	*container.Esim

	Redis *redis.Client
}

//nolint:unused,varcheck,deadcode
var infraSet = wire.NewSet(
	wire.Struct(new(Infra), "*"))

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
