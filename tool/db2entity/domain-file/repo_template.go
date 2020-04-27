package domain_file

import "github.com/jukylin/esim/pkg"

type repoTpl struct {
	Imports pkg.Imports

	StructName string

	EntityName string

	TableName string

	DelField string
}

var repoTemplate = `
package repo

{{.Imports.String}}

type {{.EntityName}}Repo interface {
	FindById(context.Context, int64) entity.{{.EntityName}}
}

type {{.StructName}} struct{

	logger log.Logger

	{{.EntityName | snakeToCamelLower | firstToLower}}Dao *dao.{{.EntityName}}Dao
}

func New{{.StructName}}(logger log.Logger) {{.EntityName}}Repo {
	{{.StructName | shorten}} := &{{.StructName}}{
		logger : logger,
	}

	if {{.StructName | shorten}}.{{.EntityName| snakeToCamelLower | firstToLower}}Dao == nil{
		{{.StructName | shorten}}.{{.EntityName| snakeToCamelLower | firstToLower}}Dao = dao.New{{.EntityName}}Dao()
	}


	return {{.StructName | shorten}}
}

func ({{.StructName | shorten}} *{{.StructName}}) FindById(ctx context.Context, id int64) entity.{{.EntityName}} {
	var {{.TableName | snakeToCamelLower}} entity.{{.EntityName}}
	var err error

	{{.TableName | snakeToCamelLower}}, err = {{.StructName | shorten}}.{{.EntityName| snakeToCamelLower | firstToLower}}Dao.Find(ctx, "*", "id = ? and {{.DelField}} = ?", id, 0)
	if err != nil {
		{{.StructName | shorten}}.logger.Errorc(ctx, err.Error())
	}

	return {{.TableName | snakeToCamelLower}}
}`

var ProvideFuncTemplate = `
func provide{{.EntityName}}Repo(esim *container.Esim) repo.{{.EntityName}}Repo {
	return repo.New{{.StructName}}(esim.Logger)
}`

func NewRepoTpl(entityName string) *repoTpl {
	rt := &repoTpl{}

	rt.EntityName = entityName
	rt.StructName = "Db" + entityName + "Repo"

	return rt
}