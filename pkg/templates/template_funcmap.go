package templates

import (
	"strings"
	"text/template"
	"unicode"

	"github.com/serenize/snaker"
)

func EsimFuncMap() map[string]interface{} {
	return template.FuncMap{
		"tolower":           strings.ToLower,
		"cutFirstToLower":   CutFirstToLower,
		"firstToLower":      FirstToLower,
		"snakeToCamel":      snaker.SnakeToCamel,
		"snakeToCamelLower": snaker.SnakeToCamelLower,
		"shorten":           Shorten,
	}
}

// Abc => a
func CutFirstToLower(s string) string {
	return strings.ToLower(string([]rune(s)[0]))
}

// Abc => abc
func FirstToLower(s string) string {
	return strings.ToLower(string([]rune(s)[0])) + string([]rune(s)[1:])
}

// Shorten shorten the string
func Shorten(s string) string {
	var result string
	var words []rune
	var max = 3

	rs := []rune(s)
	for i := 0; i < len(rs); i++ {
		if unicode.IsUpper(rs[i]) {
			words = append(words, rs[i])
		} else if i == 0 && unicode.IsLower(rs[i]) {
			words = append(words, rs[i])
		} else if string(rs[i]) == "_" || string(rs[i]) == "-" {
			if unicode.IsUpper(rs[i+1]) || unicode.IsLower(rs[i+1]) {
				words = append(words, rs[i+1])
			}
		}
	}

	if len(words) > 0 {
		for k, word := range words {
			if k >= max {
				continue
			}
			result += string(word)
		}
	} else {
		result = s[0:max]
	}

	return strings.ToLower(result)
}
