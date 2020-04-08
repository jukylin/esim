package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"strconv"
	"strings"
	"github.com/jukylin/esim/log"
	 _ "github.com/go-sql-driver/mysql"
)

type ColumnsInter interface {
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

func NewDBColumnsInter(logger log.Logger) ColumnsInter {
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

// Generate go struct entries for a map[string]interface{} structure
func generateMysqlTypes(columns []columns, depth int,
	jsonAnnotation bool, gormAnnotation bool, gureguTypes bool, v *viper.Viper) generateMysqlInfo {

	genMysqlInfo := generateMysqlInfo{}

	imports := make([]string, 0)

	var structure string

	structure += "struct {"

	var comment string
	var delKey string

	autoTime := AutoTime{}

	for _, column := range columns {
		nullable := false
		if column.IsNullAble == "YES" {
			nullable = true
		}

		primary := ""
		if column.ColumnKey == "PRI" {
			primary = ";primary_key"
		}

		col_default := ""
		if nullable == false {
			if column.ColumnDefault != "CURRENT_TIMESTAMP" && column.ColumnDefault != "" {
				col_default = ";default:'" + column.ColumnDefault + "'"
			}
		}

		// Get the corresponding go value type for this mysql type
		var valueType string
		// If the guregu (https://github.com/guregu/null) CLI option is passed use its types, otherwise use go's sql.NullX

		valueType = mysqlTypeToGoType(column.DataType, nullable, gureguTypes)
		if valueType == golangTime {
			imports = append(imports, "time")
		} else if strings.Index(valueType, "sql.") != -1 {
			imports = append(imports, "database/sql")
		} else if strings.Index(valueType, "null.") != -1 {
			imports = append(imports, "github.com/guregu/null")
		}

		if primary != "" {
			genMysqlInfo.priKeyType = valueType
		}

		fieldName := fmtFieldName(stringifyFirstChar(column.ColumnName))

		if column.ColumnDefault == "CURRENT_TIMESTAMP" {
			autoTime.CurTimeStamp = append(autoTime.CurTimeStamp, fieldName)
		}

		if column.Extra == "on update CURRENT_TIMESTAMP" {
			autoTime.OnUpdateTimeStamp = append(autoTime.OnUpdateTimeStamp, column.ColumnName)
		}

		structure += "\n\n"

		if column.ColumnComment != "" {
			structure += "//" + column.ColumnComment
		}

		var annotations []string
		if gormAnnotation == true {
			annotations = append(annotations, fmt.Sprintf("gorm:\"column:%s%s%s\"", column.ColumnName, primary, col_default))
		}

		if jsonAnnotation == true {
			annotations = append(annotations, fmt.Sprintf("json:\"%s%s,omitempty\"", column.ColumnName, primary))
		}

		if strings.Index(column.ColumnName, "del") != -1 &&
			strings.Index(column.ColumnName, "is") != -1 {
			delKey = column.ColumnName
		}

		if valueType == "string" ||
			valueType == sqlNullString || valueType == gureguNullString {
			if v.GetBool("valid") == true {
				//`validate:"max=10"`
				if column.CharacterMaximumLength != "" {
					imports = append(imports, "gopkg.in/go-playground/validator.v9")
					annotations = append(annotations, fmt.Sprintf("validate:\"max=%s\"", column.CharacterMaximumLength))
				}
			}

			if v.GetBool("mod") == true {
				//`validate:"max=10"`
				if column.CharacterMaximumLength != "" {
					//imports = append(imports, "gopkg.in/go-playground/mold.v2")
					annotations = append(annotations, fmt.Sprintf("mod:\"trim\""))
				}
			}
		}

		if len(annotations) > 0 {
			structure += fmt.Sprintf("\n%s %s `%s`",
				fieldName,
				valueType,
				strings.Join(annotations, " "))
		} else {
			structure += fmt.Sprintf("\n%s %s",
				fieldName,
				valueType)
		}
	}

	if v.GetBool("hasdata") == true {
		structure += "\n\n //用于判断是否有查询结果 \n"
		structure += "HasData bool `json:\"has_data\" gorm:\"-\"` \n"
	}

	genMysqlInfo.dbTypes = structure
	genMysqlInfo.imports = imports
	genMysqlInfo.comment = comment
	genMysqlInfo.autoTime = autoTime
	genMysqlInfo.del_key = delKey

	return genMysqlInfo
}

// mysqlTypeToGoType converts the mysql types to go compatible sql.Nullable (https://golang.org/pkg/database/sql/) types
func mysqlTypeToGoType(mysqlType string, nullable bool, gureguTypes bool) string {
	switch mysqlType {
	case "tinyint", "int", "smallint", "mediumint":
		if nullable {
			if gureguTypes {
				return gureguNullInt
			}
			return sqlNullInt
		}
		return golangInt
	case "bigint":
		if nullable {
			if gureguTypes {
				return gureguNullInt
			}
			return sqlNullInt
		}
		return golangInt64
	case "char", "enum", "varchar", "longtext", "mediumtext", "text", "tinytext":
		if nullable {
			if gureguTypes {
				return gureguNullString
			}
			return sqlNullString
		}
		return "string"
	case "date", "datetime", "time", "timestamp":
		if nullable && gureguTypes {
			return gureguNullTime
		}
		return golangTime
	case "decimal", "double":
		if nullable {
			if gureguTypes {
				return gureguNullFloat
			}
			return sqlNullFloat
		}
		return golangFloat64
	case "float":
		if nullable {
			if gureguTypes {
				return gureguNullFloat
			}
			return sqlNullFloat
		}
		return golangFloat32
	case "binary", "blob", "longblob", "mediumblob", "varbinary":
		return golangByteArray
	}
	return ""
}
