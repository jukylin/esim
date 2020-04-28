package factory

import (
	"bytes"
	"text/template"
	"github.com/jukylin/esim/pkg/templates"
)

type Plural struct {
	PluralName string

	StructName string

	Star string
}

func NewPlural() Plural {
	return Plural{}
}

var pluralNewTemplate = `func New{{.PluralName}}() *{{.PluralName}} {
	{{.PluralName | snakeToCamelLower}} := {{.PluralName | snakeToCamelLower | firstToLower}}Pool.Get().(*{{.PluralName}})
	return {{.PluralName | snakeToCamelLower}}
}
`

var pluralreleaseTemplate = `func ({{.PluralName | snakeToCamelLower | shorten}} *{{.PluralName}}) Release() {
	*{{.PluralName | snakeToCamelLower | shorten}} = (*{{.PluralName | snakeToCamelLower | shorten}})[:0]
	{{.PluralName | snakeToCamelLower | firstToLower}}Pool.Put({{.PluralName | snakeToCamelLower | shorten}})
}
`

var pluralTypeTemplate = `type {{.PluralName}} []{{.Star}}{{.StructName}}`


func (pl Plural) NewString() string {
	tmpl, err := template.New("plural_new_template").Funcs(templates.EsimFuncMap()).
		Parse(pluralNewTemplate)
	if err != nil{
		panic(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, pl)
	if err != nil{
		panic(err.Error())
	}

	return buf.String()
}


func (pl Plural) ReleaseString() string {
	tmpl, err := template.New("plural_release_template").Funcs(templates.EsimFuncMap()).
		Parse(pluralreleaseTemplate)
	if err != nil{
		panic(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, pl)
	if err != nil{
		panic(err.Error())
	}

	return buf.String()
}


func (pl Plural) TypeString() string {
	tmpl, err := template.New("plural_type_template").Funcs(templates.EsimFuncMap()).
		Parse(pluralTypeTemplate)
	if err != nil{
		panic(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, pl)
	if err != nil{
		panic(err.Error())
	}

	return buf.String()
}