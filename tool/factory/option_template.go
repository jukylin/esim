package factory

import (
	"bytes"
	"text/template"
)

var confOptionTemplate = `
func With{{.StructName}}Conf(conf config.Config) {{.StructName}}Option {
	return func({{.StructName|snakeToCamelLower|Shorten}} {{.ReturnVarStr}}) {
		{{.StructName|snakeToCamelLower|Shorten}}.conf = conf
	}
}
`

var loggerOptionTemplate = `
func With{{.StructName}}Logger(logger log.Logger) {{.StructName}}Option {
	return func({{.StructName|snakeToCamelLower|Shorten}} {{.ReturnVarStr}}) {
		{{.StructName|snakeToCamelLower|Shorten}}.logger = logger
	}
}
`

type optionTpl struct {
	StructName string

	ReturnVarStr string
}

func newOptionTpl() *optionTpl {
	ot := &optionTpl{}

	return ot
}

func (ot *optionTpl) confString(structName, returnVarStr string) string {
	ot.StructName = structName
	ot.ReturnVarStr = returnVarStr

	tmpl, err := template.New("conf_template").Parse(confOptionTemplate)
	if err != nil {
		panic(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, ot)
	if err != nil {
		panic(err.Error())
	}

	return buf.String()
}

func (ot *optionTpl) loggerString(structName, returnVarStr string) string {
	ot.StructName = structName
	ot.ReturnVarStr = returnVarStr

	tmpl, err := template.New("logger_template").Parse(loggerOptionTemplate)
	if err != nil {
		panic(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, ot)
	if err != nil {
		panic(err.Error())
	}

	return buf.String()
}
