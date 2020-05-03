package factory

import (
	"github.com/jukylin/esim/pkg/templates"
)

type PoolTpl struct {
	VarPoolName string

	StructName string

	tpl templates.Tpl
}

func NewPoolTpl() PoolTpl {
	pt := PoolTpl{}
	pt.tpl = templates.NewTextTpl()
	return pt
}

var poolTemplate = `{{.VarPoolName | snakeToCamelLower | firstToLower}} = sync.Pool{
	New: func() interface{} {
		return &{{.StructName}}{}
	},
}
`

func (pt PoolTpl) String() string {
	result, err := pt.tpl.Execute("pool_template", poolTemplate, pt)

	if err != nil {
		panic(err.Error())
	}

	return result
}
