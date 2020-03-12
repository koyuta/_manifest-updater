package repository

import "context"

type Repository interface {
	PushReplaceTagCommit(context.Context, string) error
	CreatePullRequest(context.Context) error
}
