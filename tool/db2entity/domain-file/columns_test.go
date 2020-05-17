package domain_file

import (
	"testing"

	"github.com/jukylin/esim/log"
	"github.com/stretchr/testify/assert"
	//"github.com/davecgh/go-spew/spew"
)

func TestDBColumnsInter_GetColumns(t *testing.T) {
	logger := log.NewLogger()
	dbcColumns := NewDBColumnsInter(logger)
	dbConf := &DbConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "123456",
		Database: "test_1",
		Table:    "test",
	}
	_, err := dbcColumns.SelectColumns(dbConf)
	assert.Nil(t, err)
	// spew.Dump(columns)
}

func TestColumns_IsEntity(t *testing.T) {
	cs := Columns{}

	assert.False(t, cs.IsEntity())

	cs = append(cs, Column{ColumnKey : "PRI"})
	assert.True(t, cs.IsEntity())
}