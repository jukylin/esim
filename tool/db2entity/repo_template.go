package db2entity

import "github.com/jukylin/esim/pkg"

type repoTmp struct {
	Imports pkg.Imports

	StructName string

	TableName string

	DelField string
}

var repoTemplate = `
package repo

{{.Imports.String}}

type {{.StructName}}Repo interface {
	FindById(context.Context, int64) entity.{{.StructName}}
}

type DB{{.StructName}}Repo struct{

	logger log.Logger

	{{.StructName| tolower}}Dao *dao.{{.StructName}}Dao
}

func NewDB{{.StructName}}Repo(logger log.Logger) {{.StructName}}Repo {
	repo := &DB{{.StructName}}Repo{
		logger : logger,
	}

	if repo.{{.StructName| tolower}}Dao == nil{
		repo.{{.StructName| tolower}}Dao = dao.New{{.StructName}}Dao()
	}


	return repo
}

func (this *DB{{.StructName}}Repo) FindById(ctx context.Context, id int64) entity.{{.StructName}} {
	var {{.TableName}} entity.{{.StructName}}
	var err error

	{{.TableName}}, err = this.{{.StructName| tolower}}Dao.Find(ctx, "*", "id = ? and {{.DelField}} = ?", id, 0)

	return {{.TableName}}
}

`
