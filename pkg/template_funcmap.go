package pkg

import (
	"text/template"
	"strings"
)

func EsimFuncMap() template.FuncMap {
	return template.FuncMap{
		"tolower" : strings.ToLower,
		"cutFirstToLower" : CutFirstToLower,
		"firstToLower" : FirstToLower,
		"snakeToCamel" : SnakeToCamel,
	}
}

//Abc => a
func CutFirstToLower(s string) string {
	return strings.ToLower(string([]rune(s)[0]))
}

//Abc => abc
func FirstToLower(s string) string {
	return strings.ToLower(string([]rune(s)[0])) + string([]rune(s)[1:])
}


func SnakeToCamel(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}