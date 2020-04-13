package db2entity

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/spf13/viper"
)


func TestDb2Entity_Run(t *testing.T) {


	db2EntityOptions := Db2EntityOptions{}

	db2Entity := NewDb2Entity(db2EntityOptions.WithLogger(log.NewLogger()),
		db2EntityOptions.WithColumnsInter(),
		db2EntityOptions.WithIfaceWrite(file_dir.EsimWriter{}))

	v := viper.New()
	v.Set("entity_target", "./example")
	v.Set("dao_target", "./example")
	v.Set("repo_target", "./example")
	v.Set("database", "user")
	v.Set("table", "test")

	db2Entity.Run(v)
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