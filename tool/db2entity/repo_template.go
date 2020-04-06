package db2entity


var repoTemplate = `
package repo

import (
	"context"
	"{{.CurrentDir}}/internal/domain/{{.Boubctx}}entity"
	"{{.CurrentDir}}/internal/infra/dao"
	"github.com/jukylin/esim/log"
)


type {{.StructName}}Repo interface {
	FindById(context.Context, int64) entity.{{.StructName}}
}

type DB{{.StructName}}Repo struct{

	logger log.Logger

	{{.StructName| firstToLower}}Dao *dao.{{.StructName}}Dao
}

func NewDB{{.StructName}}Repo(logger log.Logger) {{.StructName}}Repo {
	repo := &DB{{.StructName}}Repo{
		logger : logger,
	}

	if repo.{{.StructName| firstToLower}}Dao == nil{
		repo.{{.StructName| firstToLower}}Dao = dao.New{{.StructName}}Dao()
	}


	return repo
}

func (this *DB{{.StructName}}Repo) FindById(ctx context.Context, id int64) entity.{{.StructName}} {
	var {{.TableName}} entity.{{.StructName}}
	var err error

	{{.TableName}}, err = this.{{.StructName| firstToLower}}Dao.Find(ctx, "*", "id = ? ", id)

	if err != nil{
		this.logger.Errorf(err.Error())
		return {{.TableName}}
	}

	return {{.TableName}}
}

`
