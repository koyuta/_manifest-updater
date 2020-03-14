package updater

import (
	"context"

	"github.com/koyuta/manifest-updater/pkg/registry"
	"github.com/koyuta/manifest-updater/pkg/repository"
)

type Updater struct {
	Registry   registry.Registry
	Repository repository.Repository
}

func NewUpdater(regi registry.Registry, repo repository.Repository) *Updater {
	return &Updater{Registry: regi, Repository: repo}
}

func (u *Updater) Run(ctx context.Context) error {
	tag, err := u.Registry.FetchLatestTag(ctx)
	if err != nil {
		return err
	}
	if err := u.Repository.PushReplaceTagCommit(ctx, tag); err != nil {
		return err
	}
	return u.Repository.CreatePullRequest(ctx)
}