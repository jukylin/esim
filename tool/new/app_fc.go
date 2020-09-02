package new

var (
	appfc = &FileContent{
		FileName: "user.go",
		Dir:      "internal/application",
		Content: `package application

import (
	"context"
	"{{.ProPath}}{{.ServerName}}/internal/infra"
	"{{.ProPath}}{{.ServerName}}/internal/domain/user/entity"
)

type UserService struct {
	*infra.Infra
}

func NewUserSvc(infraer *infra.Infra)*UserService{
	svc := &UserService{infraer}

	return svc
}

func (svc *UserService) GetUserInfo(ctx context.Context, username string) (user entity.User) {
	user = svc.UserRepo.FindByUserName(ctx, username)

	return
}`,
	}
)

func initAppFiles() {
	Files = append(Files, appfc)
}
