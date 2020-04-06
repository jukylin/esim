package pkg

import (
	"text/template"
	"strings"
)

func EsimFuncMap() template.FuncMap {
	return template.FuncMap{
		"tolower" : strings.ToLower,
	}
}