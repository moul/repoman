package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"go.uber.org/zap"
	"moul.io/u"
)

type project struct {
	Path string
	Git  struct {
		Root          string
		MainBranch    string
		CurrentBranch string
		OriginRemotes []string

		head   *plumbing.Reference
		repo   *git.Repository
		origin *git.Remote
	}
}

// nolint:nestif
func projectFromPath(path string) (*project, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("incorrect path: %q: %w", path, err)
	}
	path = abs

	if !u.DirExists(path) {
		return nil, fmt.Errorf("path is not a directory: %q", path) // nolint:goerr113
	}

	project := &project{Path: path}
	project.Git.Root = gitFindRootDir(path)
	if project.Git.Root != "" {
		// open local repo
		{
			repo, err := git.PlainOpen(project.Git.Root)
			if err != nil {
				return nil, fmt.Errorf("failed to open git repo: %q: %w", project.Git.Root, err)
			}
			project.Git.repo = repo
		}

		// current branch
		{
			head, err := project.Git.repo.Head()
			if err != nil {
				return nil, fmt.Errorf("failed to get HEAD: %w", err)
			}
			project.Git.head = head
			project.Git.CurrentBranch = project.Git.head.Name().Short()
		}

		// 'origin' remote
		{
			origin, err := project.Git.repo.Remote("origin")
			if err != nil {
				return nil, fmt.Errorf("failed to get 'origin' remote: %w", err)
			}
			project.Git.origin = origin
			project.Git.OriginRemotes = origin.Config().URLs
		}

		// main branch
		{
			ref, err := project.Git.repo.Reference("refs/remotes/origin/HEAD", true)
			if err != nil { // if it fails, we try to fetch origin and then we retry
				err = project.Git.origin.Fetch(&git.FetchOptions{
					Depth:    1,
					Progress: os.Stdout,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to fetch origin: %w", err)
				}
				ref, err = project.Git.repo.Reference("refs/remotes/origin/HEAD", true)
				if err != nil {
					return nil, fmt.Errorf("failed to get main branch: %w", err)
				}
			}
			project.Git.MainBranch = strings.TrimPrefix(ref.Name().Short(), "origin/")
		}
	} else {
		logger.Warn("project not within a git directory", zap.String("path", path))
	}

	return project, nil
}

func gitFindRootDir(path string) string {
	for {
		if u.DirExists(filepath.Join(path, ".git")) {
			return path
		}
		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}
	return ""
}
