package db2entity

import "github.com/jukylin/esim/pkg"

type daoTpl struct{
	Imports pkg.Imports

	StructName string

	DataBaseName string

	TableName string

	PriKeyType string
}

var daoTemplate = `package dao

{{.Imports.String}}

type {{.StructName}}Dao struct{
	mysql *mysql.MysqlClient
}

func New{{.StructName}}Dao() *{{.StructName}}Dao {
	dao := &{{.StructName}}Dao{
		mysql : mysql.NewMysqlClient(),
	}

	return dao
}


//主库
func (this *{{.StructName}}Dao) GetDb(ctx context.Context) *gorm.DB  {
	return this.mysql.GetCtxDb(ctx, "{{.DataBaseName}}").Table("{{.TableName}}")
}

//从库
func (this *{{.StructName}}Dao) GetSlaveDb(ctx context.Context) *gorm.DB  {
	return this.mysql.GetCtxDb(ctx, "{{.DataBaseName}}_slave").Table("{{.TableName}}")
}


//返回 自增id，错误
func (this *{{.StructName}}Dao) Create(ctx context.Context, {{.StructName| firstToLower}} *entity.{{.StructName}}) ({{.PriKeyType}}, error){
	db := this.GetDb(ctx).Create({{.StructName| firstToLower}})
	if db.Error != nil{
		return {{.PriKeyType}}(0), db.Error
	}else{
		return {{.PriKeyType}}({{.StructName| firstToLower}}.ID), nil
	}
}

//ctx, "name = ?", "test"
func (this *{{.StructName}}Dao) Count(ctx context.Context, query interface{}, args ...interface{}) (int64, error){
	var count int64
	db := this.GetSlaveDb(ctx).Where(query, args...).Count(&count)
	if db.Error != nil{
		return count, db.Error
	}else{
		return count, nil
	}
}

// ctx, "id,name", "name = ?", "test"
func (this *{{.StructName}}Dao) Find(ctx context.Context, squery , wquery interface{}, args ...interface{}) (entity.{{.StructName}}, error){
	var {{.StructName| firstToLower}} entity.{{.StructName}}
	db := this.GetSlaveDb(ctx).Select(squery).
		Where(wquery, args...).First(&{{.StructName| firstToLower}})
	if db.Error != nil{
		return {{.StructName| firstToLower}}, db.Error
	}else{
		return {{.StructName| firstToLower}}, nil
	}
}


// ctx, "id,name", "name = ?", "test"
//最多取10条
func (this *{{.StructName}}Dao) List(ctx context.Context, squery , wquery interface{}, args ...interface{}) ([]entity.{{.StructName}}, error){
	{{.StructName| firstToLower}}s := []entity.{{.StructName}}{}
	db := this.GetSlaveDb(ctx).Select(squery).
		Where(wquery, args...).Limit(10).Find(&{{.StructName| firstToLower}}s)
	if db.Error != nil{
		return {{.StructName| firstToLower}}s, db.Error
	}else{
		return {{.StructName| firstToLower}}s, nil
	}
}

func (this *{{.StructName}}Dao) DelById(ctx context.Context, id {{.PriKeyType}}) (bool, error){
	var del{{.StructName}} entity.{{.StructName}}

	if del{{.StructName}}.DelKey() == ""{
		return false, errors.New("找不到 is_del / is_deleted / is_delete 字段")
	}

	del{{.StructName}}.ID = id
	db := this.GetDb(ctx).Update(map[string]interface{}{del{{.StructName}}.DelKey(): 1})
	if db.Error != nil{
		return false, db.Error
	}else{
		return true, nil
	}
}

//ctx, map[string]interface{}{"name": "hello"}, "name = ?", "test"
//返回影响数
func (this *{{.StructName}}Dao) Update(ctx context.Context, update map[string]interface{}, query interface{}, args ...interface{}) (int64, error) {
	db := this.GetDb(ctx).Where(query, args).
		Updates(update)
	return db.RowsAffected, db.Error
}
`
