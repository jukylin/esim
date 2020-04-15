package db2entity

var (
	cols         = make([]columns, 0)
	infraContent string
)

func init() {
	col1 := columns{
		ColumnName:    "user_name",
		DataType:      "varchar",
		IsNullAble:    "YES",
		ColumnComment: "user name",
	}
	cols = append(cols, col1)

	col2 := columns{
		ColumnName: "id",
		ColumnKey:  "PRI",
		DataType:   "int",
		IsNullAble: "NO",
	}
	cols = append(cols, col2)

	col3 := columns{
		ColumnName: "update_time",
		DataType:   "timestamp",
		IsNullAble: "NO",
		Extra:      "on update CURRENT_TIMESTAMP",
	}
	cols = append(cols, col3)

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
}
