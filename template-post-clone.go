package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"moul.io/u"
)

func doTemplatePostClone(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}
	paths := u.UniqueStrings(args)
	g, ctx := errgroup.WithContext(ctx)
	logger.Debug("doTemplatePostClone", zap.Any("opts", opts), zap.Strings("projects", paths))

	var errs error

	for _, path := range paths {
		path := path
		g.Go(func() error {
			err := doTemplatePostCloneOnce(ctx, path)
			if err != nil {
				errs = multierr.Append(errs, fmt.Errorf("%q: %w", path, err))
			}
			return nil
		})
	}
	_ = g.Wait()
	return errs
}

// nolint:gocognit,nestif
func doTemplatePostCloneOnce(_ context.Context, path string) error {
	project, err := projectFromPath(path)
	if err != nil {
		return fmt.Errorf("invalid project: %w", err)
	}

	// prepare workspace
	{
		err := project.prepareWorkspace(opts.TemplatePostClone.Project)
		if err != nil {
			return fmt.Errorf("prepare workspace: %w", err)
		}
	}

	// rm go binary
	if opts.TemplatePostClone.RemoveGoBinary {
		logger.Debug("remove go binary", zap.String("project", project.Path))
		// git rm main*.go
		err := project.Git.workTree.RemoveGlob("main*.go")
		if err != nil {
			return fmt.Errorf("rm main*.go: %w", err)
		}

		// patch Makefile
		{
			path := filepath.Join(project.Git.Root, "Makefile")
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read Makefile: %w", err)
			}
			content = regexp.MustCompile(`(?m)^(DOCKER_IMAGE|GOBINS|NPM_PACKAGES) .=.*\n`).ReplaceAll(content, []byte(""))
			content = regexp.MustCompile(`(?ms)^generate:.*.PHONY: generate\n\n`).ReplaceAll(content, []byte(""))
			err = ioutil.WriteFile(path, content, 0)
			if err != nil {
				return fmt.Errorf("write file: %q: %w", path, err)
			}
			if _, err := project.Git.workTree.Add("Makefile"); err != nil {
				return fmt.Errorf("git add %q: %w", path, err)
			}
		}

		// remove files
		for _, filename := range []string{"Dockerfile", ".goreleaser.yml", ".github/workflows/docker.yml"} {
			_, err := project.Git.workTree.Remove(filename)
			if err != nil {
				return fmt.Errorf("git rm %q: %w", filename, err)
			}
		}

		// perform various tasks
		{
			script := `
			main() {
				make generate go.depaware-update
				git add AUTHORS README.md depaware.txt
				make tidy
				git add go.mod go.sum
				git status
			}
			main
		`
			cmd := exec.Command("/bin/sh", "-xec", script)
			cmd.Stdout = os.Stderr
			cmd.Stderr = os.Stderr
			cmd.Dir = project.Path
			cmd.Env = os.Environ()

			err := cmd.Run()
			if err != nil {
				return fmt.Errorf("standard script execution failed: %w", err)
			}
		}
	}

	// find and replace
	{
		logger.Debug("patch files to remove template strings", zap.String("project", project.Path))
		visit := func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("walk dir: %q: %w", path, err)
			}
			if info.IsDir() && info.Name() == ".git" {
				return filepath.SkipDir
			}
			if info.IsDir() {
				return nil
			}
			// FIXME: ignore binary files
			// FIXME: ignore huge files

			// find & replace per file
			{
				content, err := ioutil.ReadFile(path)
				if err != nil {
					return fmt.Errorf("read file: %q: %w", path, err)
				}
				newContent := strings.ReplaceAll(string(content), opts.TemplatePostClone.TemplateName, project.Git.RepoName)
				newContent = strings.ReplaceAll(newContent, opts.TemplatePostClone.TemplateOwner, project.Git.RepoOwner)
				if string(content) != newContent {
					logger.Debug("patch file", zap.String("path", path))
					err = ioutil.WriteFile(path, []byte(newContent), 0)
					if err != nil {
						return fmt.Errorf("write file: %q: %w", path, err)
					}
				}
			}

			return nil
		}
		if err := filepath.Walk(project.Path, visit); err != nil {
			return fmt.Errorf("walk project's dir: %w", err)
		}
	}

	// push changes
	{
		err := project.pushChanges(opts.TemplatePostClone.Project, "dev/moul/template-post-clone", "chore: template post clone 🤖")
		if err != nil {
			return fmt.Errorf("push changes: %w", err)
		}
	}
	return nil
}
