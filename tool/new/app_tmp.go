package new

func ServiceInit() {
	fc1 := &FileContent{
		FileName: "user.go",
		Dir:      "internal/application",
		Content: `package application

import (
	"context"
	"{{PROPATH}}{{service_name}}/internal/infra"
	"{{PROPATH}}{{service_name}}/internal/domain/user/entity"
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
