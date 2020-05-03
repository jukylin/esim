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
	//spew.Dump(columns)
}
