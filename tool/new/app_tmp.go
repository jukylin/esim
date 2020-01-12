package new

func ServiceInit() {
	fc1 := &FileContent{
		FileName: "user.go",
		Dir:      "internal/app",
		Content: `package app

import (
	"context"
	"{{PROPATH}}{{service_name}}/internal/infra"
	"{{PROPATH}}{{service_name}}/internal/domain/entity"
)

type UserService struct {
	*infra.Infra
}

func NewUserSvc(infra *infra.Infra)*UserService{
	svc := &UserService{infra}

	return svc
}

func (svc *UserService) GetUserInfo(ctx context.Context, username string) (user entity.User) {

	user = svc.UserRepo.FindByUserName(ctx, username)

	return
}
`,
	}

	Files = append(Files, fc1)
}
