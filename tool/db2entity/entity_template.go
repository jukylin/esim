package db2entity

import "github.com/jukylin/esim/pkg"

type entityTmp struct {
	Imports pkg.Imports

	Fields pkg.Fields

	StructName string

	CreateTime string

	LastUpdateTime string

	LastUpdateTimeStr string

	DelField string
}

var entityTemplate = `package entity

{{.Imports.String}}

type {{.StructName}} struct {
{{.Fields.String}}
}

// delete field
func (c *{{.StructName}}) DelKey() string {
	return "{{.DelField}}"
}

{{if or (.CreateTime) (.LastUpdateTime)}}
//自动增加时间
func (this *{{.StructName}}) BeforeCreate(scope *gorm.Scope) (err error) {

	switch scope.Value.(type) {
	case *{{.StructName}}:

		val := scope.Value.(*{{.StructName}})

		{{if .CreateTime}}
		if val.{{.CreateTime}}.Unix() < 0 {
			val.{{.CreateTime}} = time.Now()
		}
		{{end}}

		{{if .LastUpdateTime}}
		if val.{{.LastUpdateTime}}.Unix() < 0 {
			val.{{.LastUpdateTime}} = time.Now()
		}
		{{end}}
	}

	return
}
{{end}}

{{if .LastUpdateTimeStr }}
//自动添加更新时间  没有trim
func (this *{{.StructName}}) BeforeSave(scope *gorm.Scope) (err error) {
	val, ok := scope.InstanceGet("gorm:update_attrs")
	if ok {
		switch val.(type) {
		case map[string]interface{}:
			mapVal := val.(map[string]interface{})
			if _, ok := mapVal["{{.LastUpdateTimeStr}}"]; !ok {
				mapVal["{{.LastUpdateTimeStr}}"] = time.Now()
			}
		}
	}
	return
}
{{end}}
`





