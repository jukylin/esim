
package repo

import (

 "context"
 "github.com/jukylin/esim/log"
 "github.com/jukylin/esim/tool/db2entity/./example/entity"
 "github.com/jukylin/esim/tool/db2entity/example/dao"
)

type TestRepo interface {
	FindById(context.Context, int64) entity.Test
}

type DBTestRepo struct{

	logger log.Logger

	testDao *dao.TestDao
}

func NewDBTestRepo(logger log.Logger) TestRepo {
	repo := &DBTestRepo{
		logger : logger,
	}

	if repo.testDao == nil{
		repo.testDao = dao.NewTestDao()
	}


	return repo
}

func (this *DBTestRepo) FindById(ctx context.Context, id int64) entity.Test {
	var test entity.Test

	test, err = this.testDao.Find(ctx, "*", "id = ? and  = ?", id, 0)

	return test
}