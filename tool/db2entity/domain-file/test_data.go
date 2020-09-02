package domainfile

const (
	target = "example"

	database = "test"

	testStructName = "Test"

	testTable = "test"

	userStructName = "User"

	userTable = "user"

	boubctx = "boubctx"

	delField = "is_del"
)

var (
	Cols = []Column{
		{
			ColumnName:    "user_name",
			DataType:      "varchar",
			IsNullAble:    yesNull,
			ColumnComment: "user name",
		},
		{
			ColumnName: "id",
			ColumnKey:  pri,
			DataType:   "int",
			IsNullAble: noNull,
		},
		{
			ColumnName: "update_time",
			DataType:   "timestamp",
			IsNullAble: noNull,
			Extra:      upCurrentTimestamp,
		},
	}
)
