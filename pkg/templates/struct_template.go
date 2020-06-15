package templates

import (
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
)

type StructInfo struct {
	StructName string

	Fields pkg.Fields

	tpl Tpl

	logger log.Logger
}

type Option func(c *StructInfo)

func NewStructInfo(options ...Option) *StructInfo {
	structInfo := &StructInfo{}

	for _, option := range options {
		option(structInfo)
	}

	if structInfo.logger == nil {
		structInfo.logger = log.NewLogger()
	}

	if structInfo.tpl == nil {
		structInfo.tpl = NewTextTpl()
	}

	return structInfo
}

func WithTpl(tpl Tpl) Option {
	return func(si *StructInfo) {
		si.tpl = tpl
	}
}

func WithLogger(logger log.Logger) Option {
	return func(si *StructInfo) {
		si.logger = logger
	}
}

var StructTemplate = `type {{.StructName}} struct{
{{.Fields.String}}
}`

func (si *StructInfo) String() string {
	if si.StructName == "" {
		return ""
	}

	content, err := si.tpl.Execute("struct_template", StructTemplate, si)
	if err != nil {
		si.logger.Panicf(err.Error())
	}

	return content
}
