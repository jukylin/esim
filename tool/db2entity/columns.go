package db2entity

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/jukylin/esim/log"
	"strconv"
)

// Constants for return types of golang
const (
	golangByteArray  = "[]byte"
	gureguNullInt    = "null.Int"
	sqlNullInt       = "sql.NullInt64"
	golangInt        = "int"
	golangInt64      = "int64"
	gureguNullFloat  = "null.Float"
	sqlNullFloat     = "sql.NullFloat64"
	golangFloat      = "float"
	golangFloat32    = "float32"
	golangFloat64    = "float64"
	gureguNullString = "null.String"
	sqlNullString    = "sql.NullString"
	gureguNullTime   = "null.Time"
	golangTime       = "time.Time"
)

type ColumnsRepo interface {
	GetColumns(dbConf dbConfig) ([]columns, error)
}

type columns struct {
	ColumnName             string `gorm:"column:COLUMN_NAME"`
	ColumnKey              string `gorm:"column:COLUMN_KEY"`
	DataType               string `gorm:"column:DATA_TYPE"`
	IsNullAble             string `gorm:"column:IS_NULLABLE"`
	ColumnDefault          string `gorm:"column:COLUMN_DEFAULT"`
	CharacterMaximumLength string `gorm:"column:CHARACTER_MAXIMUM_LENGTH"`
	ColumnComment          string `gorm:"column:COLUMN_COMMENT"`
	Extra                  string `gorm:"column:EXTRA"`
}

type AutoTime struct {
	CurTimeStamp      []string
	OnUpdateTimeStamp []string
}

type DBColumnsInter struct {
	logger log.Logger
}

func NewDBColumnsInter(logger log.Logger) ColumnsRepo {
	dBColumnsInter := &DBColumnsInter{}
	dBColumnsInter.logger = logger
	return dBColumnsInter
}

// GetColumns Select column details
func (dc *DBColumnsInter) GetColumns(dbConf dbConfig) ([]columns, error) {

	var err error
	var db *gorm.DB
	if dbConf.password != "" {
		db, err = gorm.Open("mysql", dbConf.user+":"+dbConf.password+
			"@tcp("+dbConf.host+":"+strconv.Itoa(dbConf.port)+")/"+dbConf.database+"?&parseTime=True")
	} else {
		db, err = gorm.Open("mysql", dbConf.user+"@tcp("+dbConf.host+":"+
			strconv.Itoa(dbConf.port)+")/"+dbConf.database+"?&parseTime=True")
	}
	defer db.Close()

	if err != nil {
		dc.logger.Panicf("Open mysql err: %s", err.Error())
	}

	if db.HasTable(dbConf.table) == false {
		dc.logger.Panicf("%s 表不存在", dbConf.table)
	}

	sql := "SELECT COLUMN_NAME, COLUMN_KEY, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, " +
		" CHARACTER_MAXIMUM_LENGTH, COLUMN_COMMENT, EXTRA " +
		"FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = ? AND table_name = ?"

	cs := make([]columns, 0)

	db.Raw(sql, dbConf.database, dbConf.table).Scan(&cs)

	if err != nil {
		dc.logger.Panicf(err.Error())
	}

	return cs, nil
}
