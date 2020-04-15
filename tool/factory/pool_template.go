package factory

import (
	"bytes"
	"text/template"
)

type PoolTpl struct {
	VarPoolName string

	StructName string
}

func NewPoolTpl() PoolTpl {
	return PoolTpl{}
}

var poolTemplate = `{{.VarPoolName}} = sync.Pool{
	New: func() interface{} {
		return &{{.StructName}}{}
	},
}
`

func (pt PoolTpl) String() string {

	tmpl, err := template.New("pool_template").Parse(poolTemplate)
	if err != nil{
		panic(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {PoolTpl}{pt})
	if err != nil{
		panic(err.Error())
	}

	return buf.String()
}