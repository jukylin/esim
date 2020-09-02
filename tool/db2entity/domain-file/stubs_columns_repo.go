package domainfile

type StubsColumnsRepo struct{}

func (scr StubsColumnsRepo) SelectColumns(dbConf *DbConfig) (Columns, error) {
	var r1 error

	return Cols, r1
}
