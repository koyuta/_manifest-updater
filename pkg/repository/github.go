package repository

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/github"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
)

var BranchName = "feature/update-tag"

type GitHubRepository struct {
	URL       string
	Branch    string
	Path      string
	ImageName string
}

func NewGitHubRepository(u, b, p, i string) *GitHubRepository {
	return &GitHubRepository{URL: u, Branch: b, Path: p, ImageName: i}
}

func (g *GitHubRepository) PushReplaceTagCommit(ctx context.Context, tag string) error {
	clonepath, err := ioutil.TempDir(os.TempDir(), "_repository")
	if err != nil {
		return err
	}

	repository, err := git.PlainCloneContext(ctx, clonepath, false, &git.CloneOptions{
		URL:           g.URL,
		SingleBranch:  true,
		ReferenceName: plumbing.NewBranchReferenceName(g.Branch),
	})
	if err != nil {
		return err
	}
	worktree, err := repository.Worktree()
	if err != nil {
		return err
	}

	if err := worktree.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: plumbing.NewBranchReferenceName(BranchName),
	}); err != nil {
		return err
	}

	var paths = []string{}
	if err = filepath.Walk(
		filepath.Join(clonepath, g.Path),
		func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				paths = append(paths, path)
			}
			return nil
		}); err != nil {
		return err
	}

	re := regexp.MustCompile(fmt.Sprintf(`%s: *(?P<tag>\w[\w-\.]+)`, g.ImageName))
	for _, p := range paths {
		content, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}
		replacedContent := re.ReplaceAll(content, []byte(tag))
		if err := ioutil.WriteFile(p, replacedContent, 0); err != nil {
			return err
		}
		if _, err := worktree.Add(p); err != nil {
			return err
		}
	}

	msg := ":up: Update image tag names from manifests"
	if _, err := worktree.Commit(msg, &git.CommitOptions{
		All: true,
		Committer: &object.Signature{
			Name: "manifest-updater",
		},
	}); err != nil {
		return err
	}

	return repository.PushContext(ctx, &git.PushOptions{})
}

func (g *GitHubRepository) CreatePullRequest(ctx context.Context) error {
	endpoint, err := transport.NewEndpoint(g.URL)
	if err != nil {
		return err
	}
	path := strings.Split(endpoint.Path, "/")
	owner, reponame := path[0], path[1]

	client := github.NewClient(nil)
	_, _, err = client.PullRequests.Create(
		ctx, owner, reponame, &github.NewPullRequest{
			Title:               github.String("Automaticaly update image tags"),
			Head:                github.String(BranchName),
			Base:                github.String(g.Branch),
			Body:                github.String(""),
			MaintainerCanModify: github.Bool(true),
		},
	)
	return err
}

type GitHubEnterpriseRepository struct {
}
