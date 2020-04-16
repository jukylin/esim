package new

func init()  {
	Files = append(Files, appfc)
}

var (
	appfc = &FileContent{
	FileName: "user_service.go",
	Dir:      "internal/application",
	Content: `package application

import (
	"context"
	"{{.ProPath}}{{.ServiceName}}/internal/infra"
	"{{.ProPath}}{{.ServiceName}}/internal/domain/user/entity"
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
)