package factory

import (
	"github.com/jukylin/esim/pkg/templates"
)

type Plural struct {
	PluralName string

	StructName string

	Star string

	tpl templates.Tpl
}

func NewPlural() Plural {
	p := Plural{}
	p.tpl = templates.NewTextTpl()
	return p
}

//nolint:lll
var pluralNewTemplate = `func New{{.PluralName}}() *{{.PluralName}} {
	{{.PluralName | snakeToCamelLower}} := {{.PluralName | snakeToCamelLower | firstToLower}}Pool.Get().(*{{.PluralName}})
	return {{.PluralName | snakeToCamelLower}}
}
`

//nolint:lll
var pluralreleaseTemplate = `func ({{.PluralName | snakeToCamelLower | shorten}} *{{.PluralName}}) Release() {
	*{{.PluralName | snakeToCamelLower | shorten}} = (*{{.PluralName | snakeToCamelLower | shorten}})[:0]
	{{.PluralName | snakeToCamelLower | firstToLower}}Pool.Put({{.PluralName | snakeToCamelLower | shorten}})
}
`

var pluralTypeTemplate = `type {{.PluralName}} []{{.Star}}{{.StructName}}`

func (pl Plural) NewString() string {
	result, err := pl.tpl.Execute("plural_new_template", pluralNewTemplate, pl)
	if err != nil {
		panic(err.Error())
	}

	return result
}

func (pl Plural) ReleaseString() string {
	result, err := pl.tpl.Execute("plural_release_template", pluralreleaseTemplate, pl)
	if err != nil {
		panic(err.Error())
	}

	return result
}

func (pl Plural) TypeString() string {
	result, err := pl.tpl.Execute("plural_type_template", pluralTypeTemplate, pl)
	if err != nil {
		panic(err.Error())
	}

	return result
}
