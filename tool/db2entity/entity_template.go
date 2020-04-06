package db2entity


type entityTmp struct {
	Import string
}

var entityTemplate = `package entity

import (
{{ range .Import }}
{{.Name}} {{.ImportPath}}
{{end}}
)

type {{.StructName}} struct {
{{ range .Fields }}
{{ range $a := .Doc}}{{$a}}
{{end}}{{.Field}} {{.Tag}}
{{end}}
}

// delete field
func (c *{{.StructName}}) DelKey() string {
	return "{{.DelField}}"
}

//自动增加时间
func (this *{{.StructName}}) BeforeCreate(scope *gorm.Scope) (err error) {

	switch scope.Value.(type) {
	case *{{.StructName}}:
		val := scope.Value.(*Coupon)

		if val.{{.CreateTime}}.Unix() < 0 {
			val.{{.CreateTime}} = time.Now()
		}

		if val.{{.LastUpdateTime}}.Unix() < 0 {
			val.{{.LastUpdateTime}} = time.Now()
		}
	}

	return
}

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

`





