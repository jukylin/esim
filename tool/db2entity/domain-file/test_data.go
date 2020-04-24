package domain_file

var (
	Cols = make([]Column, 0)
)

func init()  {
	col1 := Column{
		ColumnName:    "user_name",
		DataType:      "varchar",
		IsNullAble:    "YES",
		ColumnComment: "user name",
	}
	Cols = append(Cols, col1)

	col2 := Column{
		ColumnName: "id",
		ColumnKey:  "PRI",
		DataType:   "int",
		IsNullAble: "NO",
	}
	Cols = append(Cols, col2)

	col3 := Column{
		ColumnName: "update_time",
		DataType:   "timestamp",
		IsNullAble: "NO",
		Extra:      "on update CURRENT_TIMESTAMP",
	}
	Cols = append(Cols, col3)
}