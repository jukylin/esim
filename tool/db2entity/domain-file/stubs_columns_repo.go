package domain_file

type StubsColumnsRepo struct{}

func (scr StubsColumnsRepo) SelectColumns(dbConf *DbConfig) ([]Column, error) {
	var r1 error

	return Cols, r1
}
