package new

func init()  {
	Files = append(Files, repofc1)
}

var(
	repofc1 = &FileContent{
		FileName: "user_repo.go",
		Dir:      "internal/infra/repo",
		Content: `package repo

import (
	"context"
	"github.com/jukylin/esim/log"
	"{{PROPATH}}{{service_name}}/internal/domain/user/entity"
	"{{PROPATH}}{{service_name}}/internal/infra/dao"
)

type UserRepo interface {
	FindByUserName(context.Context, string) entity.User
}

type userRepo struct {
	log log.Logger

	userDao *dao.UserDao
}

func NewDBUserRepo(logger log.Logger) UserRepo {
	repo := &userRepo{
		log: logger,
	}

	if repo.userDao == nil {
		repo.userDao = dao.NewUserDao()
	}

	return repo
}

func (this *userRepo) FindByUserName(ctx context.Context, username string) entity.User {
	var user entity.User
	var err error

	user, err = this.userDao.Find(ctx, "*", "username = ? ", username)

	if err != nil {
		this.log.Errorf(err.Error())
		return user
	}

	return user
}
`,
	}

)
