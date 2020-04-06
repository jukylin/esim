package pkg

import (
	"text/template"
	"strings"
)

func EsimFuncMap() template.FuncMap {
	return template.FuncMap{
		"tolower" : strings.ToLower,
		"firstToLower" : func(str string) string {
			return strings.ToLower(string(str[0]))
		},
	}
}