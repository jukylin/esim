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

var pluralreleaseTemplate = `func (pl *{{.PluralName}}) Release() {
	*pl = (*pl)[:0]
	{{.PluralName | tolower}}Pool.Put(pl)
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