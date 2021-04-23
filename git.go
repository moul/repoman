package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"moul.io/u"
)

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

func gitGetMainBranch(path string) (string, error) {
	// equivalent of: git symbolic-ref refs/remotes/origin/HEAD | sed 's@^refs/remotes/origin/@@'

	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	stdout, _, err := u.ExecStandaloneOutputs(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %q: %w", cmd, err)
	}

	if len(stdout) == 0 {
		return "", fmt.Errorf("invalid command output for %q", cmd)
	}

	branch := strings.TrimSpace(strings.TrimPrefix(string(stdout), "refs/remotes/origin/"))
	if branch == "" || strings.ContainsRune(branch, '/') {
		return "", fmt.Errorf("invalid branch: %q", branch)
	}

	return branch, nil
}
