package domain_file

import (
	"strings"
	"strconv"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/jukylin/esim/log"
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
	SelectColumns(dbConf *DbConfig) ([]Column, error)
}

type Column struct {
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

// SelectColumns Select column details
func (dc *DBColumnsInter) SelectColumns(dbConf *DbConfig) ([]Column, error) {

	var err error
	var db *gorm.DB
	if dbConf.Password != "" {
		db, err = gorm.Open("mysql", dbConf.User+":"+dbConf.Password+
			"@tcp("+dbConf.Host+":"+strconv.Itoa(dbConf.Port)+")/"+dbConf.Database+"?&parseTime=True")
	} else {
		db, err = gorm.Open("mysql", dbConf.User+"@tcp("+dbConf.Host+":"+
			strconv.Itoa(dbConf.Port)+")/"+dbConf.Database+"?&parseTime=True")
	}
	defer db.Close()

	if err != nil {
		dc.logger.Panicf("Open mysql err: %s", err.Error())
	}

	if db.HasTable(dbConf.Table) == false {
		dc.logger.Panicf("%s 表不存在", dbConf.Table)
	}

	sql := "SELECT COLUMN_NAME, COLUMN_KEY, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, " +
		" CHARACTER_MAXIMUM_LENGTH, COLUMN_COMMENT, EXTRA " +
		"FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = ? AND table_name = ?"

	cs := make([]Column, 0)

	db.Raw(sql, dbConf.Database, dbConf.Table).Scan(&cs)

	if err != nil {
		dc.logger.Panicf(err.Error())
	}

	return cs, nil
}

func (c *Column) GetGoType(nullable bool) string {
	switch c.DataType {
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


func (c *Column) CheckDelField() string {
	if strings.Index(c.ColumnName, "del") != -1 &&
		strings.Index(c.ColumnName, "is") != -1 {
		return c.ColumnName
	}

	return ""
}


func (c *Column) IsTime(goType string) bool {
	if goType == golangTime {
		return true
	}

	return false
}

func (c *Column) IsCurrentTimeStamp() bool {
	if c.ColumnDefault == "CURRENT_TIMESTAMP" {
		return true
	}

	return false
}

func (c *Column) IsOnUpdate() bool {
	if c.Extra == "on update CURRENT_TIMESTAMP" {
		return true
	}

	return false
}

//filterComment filter and escaping speckial string
func (c *Column) FilterComment() string {
	if c.ColumnComment != "" {
		c.ColumnComment = strings.Replace(c.ColumnComment, "\r", "\\r", -1)
		c.ColumnComment = strings.Replace(c.ColumnComment, "\n", "\\n", -1)
	}

	return c.ColumnComment
}

//filterComment filter and escaping speckial string
func (c *Column) IsPri() bool {
	if c.ColumnKey == "PRI" {
		return true
	}

	return false
}

//GetDefCol get default tag
func (c *Column) GetDefCol() string {
	if c.ColumnDefault != "CURRENT_TIMESTAMP" && c.ColumnDefault != "" {
		return ";default:'" + c.ColumnDefault + "'"
	}

	return ""
}