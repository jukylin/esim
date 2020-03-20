package repo

import "context"

type Repo interface {
	FindById(ctx context.Context, id int64)
}
