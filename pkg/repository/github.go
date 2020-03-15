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
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

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
	repodir := g.extractRepositoryFromEndpoint(endpoint)

	clonepath, err := ioutil.TempDir(os.TempDir(), repodir)
	if err != nil {
		return err
	}

	var cloneOpts = &git.CloneOptions{
		URL:           g.URL,
		SingleBranch:  true,
		ReferenceName: plumbing.NewBranchReferenceName(g.Branch),
	}
	if pemFile := g.KeyFilePath; pemFile != "" {
		k, err := ssh.NewPublicKeysFromFile("git", pemFile, "")
		if err != nil {
			return err
		}
		cloneOpts.Auth = k
	}

	repository, err := git.PlainCloneContext(ctx, clonepath, false, cloneOpts)
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
