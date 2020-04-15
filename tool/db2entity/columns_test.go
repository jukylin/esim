package db2entity

import (
	"github.com/jukylin/esim/log"
	"github.com/stretchr/testify/assert"
	"testing"
	//"github.com/davecgh/go-spew/spew"
)

func TestDBColumnsInter_GetColumns(t *testing.T) {
	logger := log.NewLogger()
	dbcColumns := NewDBColumnsInter(logger)
	dbConf := dbConfig{
		host:     "172.16.1.71",
		port:     3306,
		user:     "root",
		password: "KeDev32109!ot5",
		database: "passport",
		table:    "user",
	}
	_, err := dbcColumns.GetColumns(dbConf)
	assert.Nil(t, err)
	//spew.Dump(columns)
}
