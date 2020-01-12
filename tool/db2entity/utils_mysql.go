package db2entity

import (
	//"errors"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"strconv"
	"strings"
)

type columns struct {
	COLUMN_NAME              string `gorm:"column:COLUMN_NAME"`
	COLUMN_KEY               string `gorm:"column:COLUMN_KEY"`
	DATA_TYPE                string `gorm:"column:DATA_TYPE"`
	IS_NULLABLE              string `gorm:"column:IS_NULLABLE"`
	COLUMN_DEFAULT           string `gorm:"column:COLUMN_DEFAULT"`
	CHARACTER_MAXIMUM_LENGTH string `gorm:"column:CHARACTER_MAXIMUM_LENGTH"`
	COLUMN_COMMENT           string `gorm:"column:COLUMN_COMMENT"`
	EXTRA                    string `gorm:"column:EXTRA"`
}

type AutoTime struct {
	CurTimeStamp      []string
	OnUpdateTimeStamp []string
}

// GetColumnsFromMysqlTable Select column details from information schema and return map of map
func GetColumnsFromMysqlTable(mariadbUser string, mariadbPassword string,
	mariadbHost string, mariadbPort int, mariadbDatabase string,
	mariadbTable string) ([]columns, error) {

	var err error
	var db *gorm.DB
	if mariadbPassword != "" {
		db, err = gorm.Open("mysql", mariadbUser+":"+mariadbPassword+"@tcp("+mariadbHost+":"+strconv.Itoa(mariadbPort)+")/"+mariadbDatabase+"?&parseTime=True")
	} else {
		db, err = gorm.Open("mysql", mariadbUser+"@tcp("+mariadbHost+":"+strconv.Itoa(mariadbPort)+")/"+mariadbDatabase+"?&parseTime=True")
	}
	defer db.Close()

	// Check for error in db, note this does not check connectivity but does check uri
	if err != nil {
		fmt.Println("Error opening mysql db: " + err.Error())
		return nil, err
	}

	if db.HasTable(mariadbTable) == false {
		return nil, errors.New(mariadbTable + " 表不存在")
	}

	// Select columnd data from INFORMATION_SCHEMA
	columnDataTypeQuery := "SELECT COLUMN_NAME, COLUMN_KEY, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, " +
		" CHARACTER_MAXIMUM_LENGTH, COLUMN_COMMENT, EXTRA " +
		"FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = ? AND table_name = ?"

	if Debug {
		fmt.Println("running: " + columnDataTypeQuery)
	}

	cs := []columns{}

	db.Raw(columnDataTypeQuery, mariadbDatabase, mariadbTable).Scan(&cs)

	if err != nil {
		fmt.Println("Error selecting from db: " + err.Error())
		return nil, err
	}

	return cs, err
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
		if column.IS_NULLABLE == "YES" {
			nullable = true
		}

		primary := ""
		if column.COLUMN_KEY == "PRI" {
			primary = ";primary_key"
		}

		col_default := ""
		if nullable == false {
			if column.COLUMN_DEFAULT != "CURRENT_TIMESTAMP" && column.COLUMN_DEFAULT != "" {
				col_default = ";default:'" + column.COLUMN_DEFAULT + "'"
			}
		}

		// Get the corresponding go value type for this mysql type
		var valueType string
		// If the guregu (https://github.com/guregu/null) CLI option is passed use its types, otherwise use go's sql.NullX

		valueType = mysqlTypeToGoType(column.DATA_TYPE, nullable, gureguTypes)
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

		fieldName := fmtFieldName(stringifyFirstChar(column.COLUMN_NAME))

		if column.COLUMN_DEFAULT == "CURRENT_TIMESTAMP" {
			autoTime.CurTimeStamp = append(autoTime.CurTimeStamp, fieldName)
		}

		if column.EXTRA == "on update CURRENT_TIMESTAMP" {
			autoTime.OnUpdateTimeStamp = append(autoTime.OnUpdateTimeStamp, column.COLUMN_NAME)
		}

		structure += "\n\n //" + column.COLUMN_COMMENT

		var annotations []string
		if gormAnnotation == true {
			annotations = append(annotations, fmt.Sprintf("gorm:\"column:%s%s%s\"", column.COLUMN_NAME, primary, col_default))
		}

		if jsonAnnotation == true {
			annotations = append(annotations, fmt.Sprintf("json:\"%s%s,omitempty\"", column.COLUMN_NAME, primary))
		}

		if strings.Index(column.COLUMN_NAME, "del") != -1 &&
			strings.Index(column.COLUMN_NAME, "is") != -1 {
			delKey = column.COLUMN_NAME
		}

		if valueType == "string" ||
			valueType == sqlNullString || valueType == gureguNullString {
			if v.GetBool("valid") == true {
				//`validate:"max=10"`
				if column.CHARACTER_MAXIMUM_LENGTH != "" {
					imports = append(imports, "gopkg.in/go-playground/validator.v9")
					annotations = append(annotations, fmt.Sprintf("validate:\"max=%s\"", column.CHARACTER_MAXIMUM_LENGTH))
				}
			}

			if v.GetBool("mod") == true {
				//`validate:"max=10"`
				if column.CHARACTER_MAXIMUM_LENGTH != "" {
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
