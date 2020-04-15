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
	{{.PluralName | tolower}} := {{.PluralName | tolower}}Pool.Get().(*{{.PluralName}})
	return {{.PluralName | tolower}}
}
`

var pluralreleaseTemplate = `func (this *{{.PluralName}}) Release() {
	*this = (*this)[:0]
	{{.PluralName | tolower}}Pool.Put(this)
}
`

var pluralTypeTemplate = `type {{.PluralName}} []{{.Star}}{{.StructName}}`


func (this Plural) NewString() string {
	tmpl, err := template.New("plural_new_template").Funcs(templates.EsimFuncMap()).
		Parse(pluralNewTemplate)
	if err != nil{
		panic(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, this)
	if err != nil{
		panic(err.Error())
	}

	return buf.String()
}


func (this Plural) ReleaseString() string {
	tmpl, err := template.New("plural_release_template").Funcs(templates.EsimFuncMap()).
		Parse(pluralreleaseTemplate)
	if err != nil{
		panic(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, this)
	if err != nil{
		panic(err.Error())
	}

	return buf.String()
}


func (this Plural) TypeString() string {
	tmpl, err := template.New("plural_type_template").Funcs(templates.EsimFuncMap()).
		Parse(pluralTypeTemplate)
	if err != nil{
		panic(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, this)
	if err != nil{
		panic(err.Error())
	}

	return buf.String()
}