package dao

import (

 "context"
 "github.com/jinzhu/gorm"
 "errors"
 "github.com/jukylin/esim/mysql"
 "github.com/jukylin/esim/tool/db2entity/./example/entity"
)

type TestDao struct{
	mysql *mysql.MysqlClient
}

func NewTestDao() *TestDao {
	dao := &TestDao{
		mysql : mysql.NewMysqlClient(),
	}

	return dao
}


//主库
func (this *TestDao) GetDb(ctx context.Context) *gorm.DB  {
	return this.mysql.GetCtxDb(ctx, "user").Table("test")
}

//从库
func (this *TestDao) GetSlaveDb(ctx context.Context) *gorm.DB  {
	return this.mysql.GetCtxDb(ctx, "user_slave").Table("test")
}


//返回 自增id，错误
func (this *TestDao) Create(ctx context.Context, test *entity.Test) (int, error){
	db := this.GetDb(ctx).Create(test)
	if db.Error != nil{
		return int(0), db.Error
	}else{
		return int(test.ID), nil
	}
}

//ctx, "name = ?", "test"
func (this *TestDao) Count(ctx context.Context, query interface{}, args ...interface{}) (int64, error){
	var count int64
	db := this.GetSlaveDb(ctx).Where(query, args...).Count(&count)
	if db.Error != nil{
		return count, db.Error
	}else{
		return count, nil
	}
}

// ctx, "id,name", "name = ?", "test"
func (this *TestDao) Find(ctx context.Context, squery , wquery interface{}, args ...interface{}) (entity.Test, error){
	var test entity.Test
	db := this.GetSlaveDb(ctx).Select(squery).
		Where(wquery, args...).First(&test)
	if db.Error != nil{
		return test, db.Error
	}else{
		return test, nil
	}
}


// ctx, "id,name", "name = ?", "test"
//最多取10条
func (this *TestDao) List(ctx context.Context, squery , wquery interface{}, args ...interface{}) ([]entity.Test, error){
	tests := []entity.Test{}
	db := this.GetSlaveDb(ctx).Select(squery).
		Where(wquery, args...).Limit(10).Find(&tests)
	if db.Error != nil{
		return tests, db.Error
	}else{
		return tests, nil
	}
}

func (this *TestDao) DelById(ctx context.Context, id int) (bool, error){
	var delTest entity.Test

	if delTest.DelKey() == ""{
		return false, errors.New("找不到 is_del / is_deleted / is_delete 字段")
	}

	delTest.ID = id
	db := this.GetDb(ctx).Update(map[string]interface{}{delTest.DelKey(): 1})
	if db.Error != nil{
		return false, db.Error
	}else{
		return true, nil
	}
}

//ctx, map[string]interface{}{"name": "hello"}, "name = ?", "test"
//返回影响数
func (this *TestDao) Update(ctx context.Context, update map[string]interface{}, query interface{}, args ...interface{}) (int64, error) {
	db := this.GetDb(ctx).Where(query, args).
		Updates(update)
	return db.RowsAffected, db.Error
}
