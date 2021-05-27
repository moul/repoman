package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"moul.io/u"
)

func doMaintenance(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}
	paths := u.UniqueStrings(args)
	g, ctx := errgroup.WithContext(ctx)
	logger.Debug("doMaintenance", zap.Any("opts", opts), zap.Strings("projects", paths))

	var errs error

	for _, path := range paths {
		path := path
		g.Go(func() error {
			err := doMaintenanceOnce(ctx, path)
			if err != nil {
				errs = multierr.Append(errs, fmt.Errorf("%q: %w", path, err))
			}
			return nil
		})
	}
	_ = g.Wait()
	return errs
}

func doMaintenanceOnce(_ context.Context, path string) error {
	project, err := projectFromPath(path)
	if err != nil {
		return fmt.Errorf("invalid project: %w", err)
	}

	// prepare workspace
	{
		err := project.prepareWorkspace(opts.Maintenance.Project)
		if err != nil {
			return fmt.Errorf("prepare workspace: %w", err)
		}
	}

	// TODO
	// - repoman.yml ->
	//   - template -> moul/golang-repo-template
	//   - exclude: - README.md
	//   - no-main / lib-only
	// - auto update from template
	// - open PR / update existing one

	if opts.Maintenance.BumpDeps {
		logger.Debug("bumping deps", zap.String("project", project.Path))
		// TODO: for each dirs with a go.mod, except vendor; overridable by repoman.yml
		// TODO: overridable go binary
		cmd := exec.Command("go", "get", "-u", "./...")
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		cmd.Dir = project.Path
		cmd.Env = os.Environ()
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("exec failed: %w", err)
		}
	}

	if opts.Maintenance.Standard {
		logger.Debug("applying standard changes", zap.String("project", project.Path))
		script := `
			main() {
				# renovate.json
				mkdir -p .github
				git mv renovate.json .github/renovate.json || true
				git rm -f renovate.json || true
				cp ~/go/src/moul.io/golang-repo-template/.github/renovate.json .github/ || true
				git add .github/renovate.json || true
				git add renovate.json || true

				# dependabot
				cp ~/go/src/moul.io/golang-repo-template/.github/dependabot.yml .github/ || true
				git add .github/dependabot.yml || true

				# rules.mk
				if [ -f rules.mk ]; then cp ~/go/src/moul.io/rules.mk/rules.mk .; fi || true

				# authors
				if [ -f rules.mk ]; then make generate.authors; git add AUTHORS; fi || true

				# copyright
				set -xe; \
				for prefix in "©" "Copyright" "Copyright (c)"; do \
					for file in README.md LICENSE-APACHE LICENSE-MIT LICENSE COPYRIGHT; do \
						if [ -f "$file" ]; then \
							sed -i "s/$prefix 2014 /$prefix 2014-2021 /" $file; \
							sed -i "s/$prefix 2015 /$prefix 2015-2021 /" $file; \
							sed -i "s/$prefix 2016 /$prefix 2016-2021 /" $file; \
							sed -i "s/$prefix 2017 /$prefix 2017-2021 /" $file; \
							sed -i "s/$prefix 2018 /$prefix 2018-2021 /" $file; \
							sed -i "s/$prefix 2019 /$prefix 2019-2021 /" $file; \
							sed -i "s/$prefix 2020 /$prefix 2020-2021 /" $file; \
							sed -i "s/$prefix \([0-9][0-9][0-9][0-9]\)-20[0-9][0-9] /$prefix \1-2021 /" $file; \
							sed -i "s/$prefix 2021-2021/$prefix 2021 /" $file; \
						fi; \
					done; \
				done

				# golangci-lint fix
				sed -i "s/version: v1.26/version: v1.38/" .github/workflows/*.yml || true
				sed -i "s/version: v1.27/version: v1.38/" .github/workflows/*.yml || true
				sed -i "s/version: v1.28/version: v1.38/" .github/workflows/*.yml || true
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

	// push changes
	{
		err := project.pushChanges(opts.TemplatePostClone.Project, "dev/moul/maintenance", "chore: repo maintenance 🤖")
		if err != nil {
			return fmt.Errorf("push changes: %w", err)
		}
	}

	return nil
}
