package domain_file

import "github.com/jukylin/esim/pkg"

type daoTpl struct {
	Imports pkg.Imports

	StructName string

	EntityName string

	DataBaseName string

	TableName string

	PriKeyType string
}

//nolint:lll
var daoTemplate = `package dao

{{.Imports.String}}

type {{.StructName}} struct{
	mysql *mysql.Client
}

func New{{.StructName}}() *{{.StructName}} {
	dao := &{{.StructName}}{
		mysql : mysql.NewClient(),
	}

	return dao
}


//主库
func ({{.StructName | shorten}} *{{.StructName}}) GetDb(ctx context.Context) *gorm.DB  {
	return {{.StructName | shorten}}.mysql.GetCtxDb(ctx, "{{.DataBaseName}}").Table("{{.TableName}}")
}

//从库
func ({{.StructName | shorten}} *{{.StructName}}) GetSlaveDb(ctx context.Context) *gorm.DB  {
	return {{.StructName | shorten}}.mysql.GetCtxDb(ctx, "{{.DataBaseName}}_slave").Table("{{.TableName}}")
}


//返回 自增id，错误
func ({{.StructName | shorten}} *{{.StructName}}) Create(ctx context.Context,
		{{.EntityName| firstToLower}} *entity.{{.EntityName}}) ({{.PriKeyType}}, error){
	db := {{.StructName | shorten}}.GetDb(ctx).Create({{.EntityName| firstToLower}})
	if db.Error != nil{
		return {{.PriKeyType}}(0), db.Error
	}else{
		return {{.PriKeyType}}({{.EntityName| firstToLower}}.ID), nil
	}
}

//ctx, "name = ?", "test"
func ({{.StructName | shorten}} *{{.StructName}}) Count(ctx context.Context,
		query interface{}, args ...interface{}) (int64, error){
	var count int64
	db := {{.StructName | shorten}}.GetSlaveDb(ctx).Where(query, args...).Count(&count)
	if db.Error != nil{
		return count, db.Error
	}else{
		return count, nil
	}
}

// ctx, "id,name", "name = ?", "test"
func ({{.StructName | shorten}} *{{.StructName}}) Find(ctx context.Context, squery,
		wquery interface{}, args ...interface{}) (entity.{{.EntityName}}, error){
	var {{.EntityName| snakeToCamelLower | firstToLower}} entity.{{.EntityName}}
	db := {{.StructName | shorten}}.GetSlaveDb(ctx).Select(squery).
		Where(wquery, args...).First(&{{.EntityName| snakeToCamelLower | firstToLower}})
	if db.Error != nil{
		return {{.EntityName| snakeToCamelLower | firstToLower}}, db.Error
	}else{
		return {{.EntityName| snakeToCamelLower | firstToLower}}, nil
	}
}


// ctx, "id,name", "name = ?", "test"
//最多取10条
func ({{.StructName | shorten}} *{{.StructName}}) List(ctx context.Context, squery,
		wquery interface{}, args ...interface{}) ([]entity.{{.EntityName}}, error){
	{{.EntityName| snakeToCamelLower | firstToLower}}s := make([]entity.{{.EntityName}}, 0)
	db := {{.StructName | shorten}}.GetSlaveDb(ctx).Select(squery).
		Where(wquery, args...).Limit(10).Find(&{{.EntityName| snakeToCamelLower | firstToLower}}s)
	if db.Error != nil{
		return {{.EntityName| snakeToCamelLower | firstToLower}}s, db.Error
	}else{
		return {{.EntityName| snakeToCamelLower | firstToLower}}s, nil
	}
}

func ({{.StructName | shorten}} *{{.StructName}}) DelById(ctx context.Context,
		id {{.PriKeyType}}) (bool, error){
	var del{{.EntityName}} entity.{{.EntityName}}

	if del{{.EntityName}}.DelKey() == ""{
		return false, errors.New("找不到 is_del / is_deleted / is_delete 字段")
	}

	del{{.EntityName}}.ID = id
	db := {{.StructName | shorten}}.GetDb(ctx).Where("id = ?", id).
		Update(map[string]interface{}{del{{.EntityName}}.DelKey(): 1})
	if db.Error != nil{
		return false, db.Error
	}else{
		return true, nil
	}
}

//ctx, map[string]interface{}{"name": "hello"}, "name = ?", "test"
//返回影响数
func ({{.StructName | shorten}} *{{.StructName}}) Update(ctx context.Context,
		update map[string]interface{}, query interface{}, args ...interface{}) (int64, error) {
	db := {{.StructName | shorten}}.GetDb(ctx).Where(query, args).
		Updates(update)
	return db.RowsAffected, db.Error
}
`

func newDaoTpl(entityName string) *daoTpl {
	dt := &daoTpl{}
	dt.EntityName = entityName
	dt.StructName = entityName + "Dao"
	return dt
}
