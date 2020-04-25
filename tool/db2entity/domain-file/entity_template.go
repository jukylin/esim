package domain_file

import (
	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/pkg/templates"
)

type entityTpl struct {
	Imports pkg.Imports

	StructName string

	// CURRENT_TIMESTAMP
	CurTimeStamp []string

	//on update CURRENT_TIMESTAMP
	OnUpdateTimeStamp []string

	OnUpdateTimeStampStr []string

	DelField string

	StructInfo templates.StructInfo
}

var entityTemplate = `package entity

{{.Imports.String}}

{{.StructInfo.String}}

// delete field
func ({{.StructName | shorten}} *{{.StructName}}) DelKey() string {
	return "{{.DelField}}"
}

{{if or (.CurTimeStamp) (.OnUpdateTimeStamp)}}
//自动增加时间
func ({{.StructName | shorten}} *{{.StructName}}) BeforeCreate(scope *gorm.Scope) (err error) {

	switch scope.Value.(type) {
	case *{{.StructName}}:

		val := scope.Value.(*{{.StructName}})

		{{range $stamp := .CurTimeStamp}}
		if val.{{$stamp}}.Unix() < 0 {
			val.{{$stamp}} = time.Now()
		}
		{{end}}

		{{range $stamp := .OnUpdateTimeStamp}}
		if val.{{$stamp}}.Unix() < 0 {
			val.{{$stamp}} = time.Now()
		}
		{{end}}
	}

	return
}
{{end}}

{{if .OnUpdateTimeStampStr }}
//自动添加更新时间
func ({{.StructName | shorten}} *{{.StructName}}) BeforeSave(scope *gorm.Scope) (err error) {
	val, ok := scope.InstanceGet("gorm:update_attrs")
	if ok {
		switch val.(type) {
		case map[string]interface{}:
			mapVal := val.(map[string]interface{})
			{{range $stampStr := .OnUpdateTimeStampStr}}
			if _, ok := mapVal["{{$stampStr}}"]; !ok {
				mapVal["{{$stampStr}}"] = time.Now()
			}
			{{end}}
		}
	}
	return
}
{{end}}
`
