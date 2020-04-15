package db2entity

import (
	"github.com/jinzhu/gorm"
	"strconv"
	"github.com/jukylin/esim/log"
	 _ "github.com/go-sql-driver/mysql"
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
	ColumnName string `gorm:"column:COLUMN_NAME"`
	ColumnKey string `gorm:"column:COLUMN_KEY"`
	DataType string `gorm:"column:DATA_TYPE"`
	IsNullAble string `gorm:"column:IS_NULLABLE"`
	ColumnDefault string `gorm:"column:COLUMN_DEFAULT"`
	CharacterMaximumLength string `gorm:"column:CHARACTER_MAXIMUM_LENGTH"`
	ColumnComment string `gorm:"column:COLUMN_COMMENT"`
	Extra string `gorm:"column:EXTRA"`
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
func (this *DBColumnsInter) GetColumns(dbConf dbConfig) ([]columns, error) {

	var err error
	var db *gorm.DB
	if dbConf.password != "" {
		db, err = gorm.Open("mysql", dbConf.user + ":" + dbConf.password +
			"@tcp(" + dbConf.host + ":" + strconv.Itoa(dbConf.port)+")/" + dbConf.database + "?&parseTime=True")
	} else {
		db, err = gorm.Open("mysql", dbConf.user + "@tcp(" + dbConf.host + ":" +
			strconv.Itoa(dbConf.port) + ")/" + dbConf.database + "?&parseTime=True")
	}
	defer db.Close()

	if err != nil {
		this.logger.Panicf("Open mysql err: %s" , err.Error())
	}

	if db.HasTable(dbConf.table) == false {
		this.logger.Panicf("%s 表不存在", dbConf.table)
	}

	sql := "SELECT COLUMN_NAME, COLUMN_KEY, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, " +
		" CHARACTER_MAXIMUM_LENGTH, COLUMN_COMMENT, EXTRA " +
		"FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = ? AND table_name = ?"

	cs := []columns{}

	db.Raw(sql, dbConf.database, dbConf.table).Scan(&cs)

	if err != nil {
		this.logger.Panicf(err.Error())
	}

	return cs, nil
}

func (this *db2Entity) mysqlTypeToGoType(mysqlType string, nullable bool) string {
	switch mysqlType {
	case "tinyint", "int", "smallint", "mediumint":
		if nullable {
			return sqlNullInt
		}
		return golangInt
	case "bigint":
		if nullable {
			return sqlNullInt
		}
		return golangInt64
	case "char", "enum", "varchar", "longtext", "mediumtext", "text", "tinytext":
		if nullable {
			return sqlNullString
		}
		return "string"
	case "date", "datetime", "time", "timestamp":
		return golangTime
	case "decimal", "double":
		if nullable {
			return sqlNullFloat
		}
		return golangFloat64
	case "float":
		if nullable {
			return sqlNullFloat
		}
		return golangFloat32
	case "binary", "blob", "longblob", "mediumblob", "varbinary":
		return golangByteArray
	}
	return ""
}
