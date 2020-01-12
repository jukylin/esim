//+build wireinject

package container

import (
	"github.com/google/wire"
)

func initEsim() *Esim {
	wire.Build(esimSet)
	return nil
}

func NewMockEsim() *Esim {
	wire.Build(MockSet)
	return nil
}
