package domainfile

import (
	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/pkg/templates"
)

type entityTpl struct {
	Imports pkg.Imports

	StructName string

	// CURRENT_TIMESTAMP
	CurTimeStamp []string

	// on update CURRENT_TIMESTAMP
	OnUpdateTimeStamp []string

	OnUpdateTimeStampStr []string

	DelField string

	StructInfo *templates.StructInfo
}

var entityTemplate = `package entity

{{.Imports.String}}

{{.StructInfo.String}}

// delete field
func ({{.StructName | shorten}} *{{.StructName}}) DelKey() string {
	return "{{.DelField}}"
}
`
