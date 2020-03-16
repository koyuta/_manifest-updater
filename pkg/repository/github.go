package repository

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/github"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

var ErrTagNotReplaced = errors.New("tag not replaced")

var nowFunc = time.Now

var BranchName = "feature/update-tag"

type GitHubRepository struct {
	URL         string
	Branch      string
	Path        string
	ImageName   string
	KeyFilePath string
}

func NewGitHubRepository(u, b, p, i, k string) *GitHubRepository {
	return &GitHubRepository{
		URL:         u,
		Branch:      b,
		Path:        p,
		ImageName:   i,
		KeyFilePath: k,
	}
}

func (g *GitHubRepository) PushReplaceTagCommit(ctx context.Context, tag string) error {
	endpoint, err := transport.NewEndpoint(g.URL)
	if err != nil {
		return err
	}

	var branch = plumbing.Master
	if g.Branch != "" {
		branch = plumbing.NewBranchReferenceName(g.Branch)
	}

	clonepath := filepath.Join(
		os.TempDir(),
		g.extractOwnerFromEndpoint(endpoint),
		g.extractRepositoryFromEndpoint(endpoint),
	)

	var auth transport.AuthMethod
	if pemFile := g.KeyFilePath; pemFile != "" {
		auth, err = ssh.NewPublicKeysFromFile("git", pemFile, "")
		if err != nil {
			return err
		}
	}

	var repository *git.Repository
	if _, err := os.Stat(clonepath); os.IsNotExist(err) {
		opts := &git.CloneOptions{
			URL:           g.URL,
			SingleBranch:  true,
			ReferenceName: branch,
			Auth:          auth,
		}
		repository, err = git.PlainCloneContext(ctx, clonepath, false, opts)
		if err != nil {
			return err
		}
	} else {
		repository, err = git.PlainOpen(clonepath)
		if err != nil {
			return err
		}
	}
	worktree, err := repository.Worktree()
	if err != nil {
		return err
	}
	if err := worktree.PullContext(ctx, &git.PullOptions{
		Force:         true,
		SingleBranch:  true,
		ReferenceName: branch,
		Auth:          auth,
	}); err != nil {
		return err
	}

	checkoutOpts := &git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(BranchName),
	}
	if _, err = repository.Branch(BranchName); errors.Is(err, git.ErrBranchNotFound) {
		checkoutOpts.Create = true
	}
	if err := worktree.Checkout(checkoutOpts); err != nil {
		return err
	}

	re := regexp.MustCompile(fmt.Sprintf(`%s: *(?P<tag>\w[\w-\.]{0,127})`, g.ImageName))
	if err = filepath.Walk(
		filepath.Join(clonepath, g.Path),
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if strings.HasPrefix(path, filepath.Join(clonepath, ".git")) {
				return nil
			}

			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			replacedContent := re.ReplaceAll(content, []byte(fmt.Sprintf("%s:%s", g.ImageName, tag)))
			if err := ioutil.WriteFile(path, replacedContent, 0); err != nil {
				return err
			}
			if !bytes.Equal(content, replacedContent) {
				prefix := fmt.Sprintf("%s/", clonepath)
				if _, err := worktree.Add(strings.TrimPrefix(path, prefix)); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
		return err
	}

	status, err := worktree.Status()
	if err != nil {
		return err
	}

	// To prevent non-fast-forward error, do not commit and push
	// if no file was modified.
	if len(status) == 0 {
		return ErrTagNotReplaced
	}

	msg := ":up: Update image tag names from manifests"
	if _, err := worktree.Commit(msg, &git.CommitOptions{
		All: true,
		Author: &object.Signature{
			Name: "manifest-updater",
			When: nowFunc(),
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

	owner := g.extractOwnerFromEndpoint(endpoint)
	repoistory := g.extractRepositoryFromEndpoint(endpoint)
	pullRequest := &github.NewPullRequest{
		Title:               github.String("Automaticaly update image tags"),
		Head:                github.String(BranchName),
		Base:                github.String(g.Branch),
		Body:                github.String(""),
		MaintainerCanModify: github.Bool(true),
	}

	client := github.NewClient(nil)
	_, _, err = client.PullRequests.Create(ctx, owner, repoistory, pullRequest)
	return err
}

func (g *GitHubRepository) extractOwnerFromEndpoint(endpoint *transport.Endpoint) string {
	path := strings.Split(endpoint.Path, "/")
	owner := path[0]
	return owner
}

func (g *GitHubRepository) extractRepositoryFromEndpoint(endpoint *transport.Endpoint) string {
	path := strings.Split(endpoint.Path, "/")
	repo := path[1]
	return repo
}
