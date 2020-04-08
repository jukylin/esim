package db2entity

import (
	"testing"
	"github.com/stretchr/testify/assert"
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

func TestDb2Entity_Run(t *testing.T) {

}

func TestDb2Entity_CloumnsToEntityTmp(t *testing.T)  {

	db2Entity := &db2Entity{}

	cols := []columns{}
	col1 := columns{
		ColumnName : "user_name",
		DataType: "varchar",
		IsNullAble : "YES",
		ColumnComment : "user name",
	}
	cols = append(cols, col1)

	col2 := columns{
		ColumnName : "id",
		ColumnKey : "PRI",
		DataType: "int",
		IsNullAble : "NO",
	}
	cols = append(cols, col2)

	col3 := columns{
		ColumnName : "update_time",
		DataType: "timestamp",
		IsNullAble : "NO",
		Extra : "on update CURRENT_TIMESTAMP",
	}
	cols = append(cols, col3)

	entityTmp := db2Entity.cloumnsToEntityTmp(cols)
	assert.Equal(t, 3, len(entityTmp.Fields))
}