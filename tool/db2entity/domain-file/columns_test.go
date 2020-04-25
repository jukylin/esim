package domain_file

import (
	"github.com/jukylin/esim/log"
	"github.com/stretchr/testify/assert"
	"testing"
	//"github.com/davecgh/go-spew/spew"
)

func TestDBColumnsInter_GetColumns(t *testing.T) {
	logger := log.NewLogger()
	dbcColumns := NewDBColumnsInter(logger)
	dbConf := DbConfig{
		Host:     "127.0.0.1",
		Port:     3306,
		User:     "root",
		Password: "",
		Database: "passport",
		Table:    "user",
	}
	_, err := dbcColumns.GetColumns(dbConf)
	assert.Nil(t, err)
	//spew.Dump(columns)
}
