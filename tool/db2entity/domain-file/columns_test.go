package domainfile

import (
	"testing"

	"github.com/jukylin/esim/log"
	"github.com/stretchr/testify/assert"
	// "github.com/davecgh/go-spew/spew"
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
		Table:    testTable,
	}
	_, err := dbcColumns.SelectColumns(dbConf)
	assert.Nil(t, err)
}

func TestColumns_IsEntity(t *testing.T) {
	cs := Columns{}

	assert.False(t, cs.IsEntity())

	cs = append(cs, Column{ColumnKey: pri})
	assert.True(t, cs.IsEntity())
}

func TestColumns_FilterComment(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		want    string
	}{
		{"empty comment", "", ""},
		{"speckial comment", "abc\r\n", "abc\\r\\n"},
		{"normal comment", "abc", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := Column{}
			cl.ColumnComment = tt.comment
			assert.Equal(t, tt.want, cl.FilterComment())
		})
	}

}
