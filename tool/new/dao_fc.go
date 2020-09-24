package new

var (
	daofc1 = &FileContent{
		FileName: "user.go",
		Dir:      "internal/infra/dao",
		Content: `package dao

import (
	"context"
	"gorm.io/gorm"
	"github.com/jukylin/esim/mysql"
	"{{.ProPath}}{{.ServerName}}/internal/domain/user/entity"
)

type UserDao struct {
	mysql *mysql.Client
}

func NewUserDao() *UserDao {
	dao := &UserDao{
		mysql: mysql.NewClient(),
	}

	return dao
}

// master
func (ud *UserDao) GetDb(ctx context.Context) *gorm.DB {
	return ud.mysql.GetCtxDb(ctx, "passport").Table("user")
}

// slave
func (ud *UserDao) GetSlaveDb(ctx context.Context) *gorm.DB {
	return ud.mysql.GetCtxDb(ctx, "passport_slave").Table("user")
}


// ctx, "id,name", "name = ?", "test"
func (ud *UserDao) Find(ctx context.Context, squery, wquery interface{},
args ...interface{}) (entity.User, error) {
	var user entity.User
	user.ID = 1
	user.UserName = "demo"
	user.PassWord = "123456"

	return user, nil
}
`,
	}
)

func initDaoFiles() {
	Files = append(Files, daofc1)
}
