package new

func RepoInit() {
	fc1 := &FileContent{
		FileName: "user.go",
		Dir:      "internal/infra/repo",
		Content: `package repo

import (
	"context"
	"github.com/jukylin/esim/log"
	"{{PROPATH}}{{service_name}}/internal/domain/entity"
	"{{PROPATH}}{{service_name}}/internal/infra/repo/dao"
)

type UserRepo interface {
	FindByUserName(context.Context, string) entity.User
}

type userRepo struct {
	log log.Logger

	userDao *dao.UserDao
}

func NewUserRepo(logger log.Logger) UserRepo {
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

	Files = append(Files, fc1)
}
