package db2entity

import (
	"testing"
	//"github.com/jukylin/esim/gofmt"
)

func TestInject(t *testing.T) {

	src := `package db2entity

import (
	wire "github.com/google/wire"
	"github.com/jukylin/esim/container"
	_ "github.com/jukylin/esim"

	"sync"
)

var infrOnce sync.Once
var onceInfr *Infra

type Infra struct {

	//fggh

	*container.Esim

	//Binding
	//123
	s wire.Binding
}

var infrSet = wire.NewSet(
	wire.Struct(new(infra), "*"),

	provideEsim,
)

func NewInfr() *Infra {
	infrOnce.Do(func() {})

	return onceInfr
}

func provideEsim() *container.Esim {
	return container.NewEsim()
}
`

	src = handleInject(src, "Infra", "UserDao",
		"passport", "UserRepo", "UserDbRepo","github.com/jukylin/esim/db2entity")
	println(src)
}

//func TestGoFmt(t *testing.T)  {
//	gofmt.GoFmt("./infra.go")
//	t.Error("21231")
//}
