package factory

import (
	"github.com/jukylin/esim/pkg/templates"
)

var confOptionTemplate = `
func With{{.StructName}}Conf(conf config.Config) {{.StructName}}Option {
	return func({{.StructName|snakeToCamelLower|shorten}} {{.ReturnVarStr}}) {
		{{.StructName|snakeToCamelLower|shorten}}.conf = conf
	}
}
`

var loggerOptionTemplate = `
func With{{.StructName}}Logger(logger log.Logger) {{.StructName}}Option {
	return func({{.StructName|snakeToCamelLower|shorten}} {{.ReturnVarStr}}) {
		{{.StructName|snakeToCamelLower|shorten}}.logger = logger
	}
}
`

type optionTpl struct {
	StructName string

	ReturnVarStr string

	tpl templates.Tpl
}

func newOptionTpl(tpl templates.Tpl) *optionTpl {
	ot := &optionTpl{}
	ot.tpl = tpl

	return ot
}

func (ot *optionTpl) String(structName, returnVarStr, tplName, tpl string) string {
	ot.StructName = structName
	ot.ReturnVarStr = returnVarStr

	content, err := ot.tpl.Execute(tplName, tpl, ot)
	if err != nil {
		panic(err.Error())
	}

	return content
}
