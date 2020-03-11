package repository

type Repository interface {
	PushReplaceTagCommit(string) error
	CreatePullRequest() error
}
