package main

import (
	"fmt"
	"path/filepath"

	"go.uber.org/zap"
	"moul.io/u"
)

type project struct {
	Path string
	Git  struct {
		Root          string
		MainBranch    string
		CurrentBranch string
		OriginRemote  string
	}
}

func (p *project) checkoutMainBranch() error {
	return nil
}

func projectFromPath(path string) (*project, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("incorrect path: %q: %w", path, err)
	}
	path = abs

	if !u.DirExists(path) {
		return nil, fmt.Errorf("path is not a directory: %q", path)
	}

	project := &project{Path: path}
	project.Git.Root = gitFindRootDir(path)
	if project.Git.Root != "" {
		project.Git.MainBranch, err = gitGetMainBranch(project.Git.Root)
		if err != nil {
			return nil, fmt.Errorf("cannot guess main branch: %w", err)
		}
	} else {
		logger.Warn("project not withing a git directory", zap.String("path", path))
	}

	return project, nil
}
