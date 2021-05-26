package main

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"moul.io/u"
)

func doTemplatePostClone(ctx context.Context, args []string) error {
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

	// find and replace
	{
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
