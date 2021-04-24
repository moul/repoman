package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"moul.io/u"
)

func doMaintenance(ctx context.Context, args []string) error {
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

	if project.Git.Root == "" {
		return fmt.Errorf("not implemented: maintenance over non-git projects")
	}

	if project.Git.IsDirty {
		return fmt.Errorf("worktree is dirty, please commit or discard changes before running a maintenance") // nolint:goerr113
	}

	if !opts.Maintenance.NoFetch {
		logger.Debug("fetch origin", zap.String("project", project.Path))
		err := project.Git.origin.Fetch(&git.FetchOptions{
			Progress: os.Stderr,
		})
		switch err {
		case git.NoErrAlreadyUpToDate:
			// skip
		case nil:
			// skip
		default:
			return fmt.Errorf("failed to fetch origin: %w", err)
		}
	}

	if !project.Git.InMainBranch {
		logger.Debug("project is not using the main branch",
			zap.String("current", project.Git.CurrentBranch),
			zap.String("main", project.Git.MainBranch),
		)

		// TODO: checkout
		return fmt.Errorf("not implemented: git checkout master/main")
	}

	// - repoman.yml ->
	//   - template -> moul/golang-repo-template
	//   - exclude: - README.md
	//   - no-main / lib-only
	// - auto update from template
	// - open PR / update existing one
	//
	// _do.maintenance: _do.checkoutmaster
	//	# renovate.json
	//	mkdir -p .github
	//	git mv renovate.json .github/renovate.json || true
	//	git rm -f renovate.json || true
	//	cp ~/go/src/moul.io/golang-repo-template/.github/renovate.json .github/ || true
	//	git add .github/renovate.json || true
	//	git add renovate.json || true
	//
	//	# dependabot
	//	cp ~/go/src/moul.io/golang-repo-template/.github/dependabot.yml .github/ || true
	//	git add .github/dependabot.yml || true
	//
	//	# rules.mk
	//	if [ -f rules.mk ]; then cp ~/go/src/moul.io/rules.mk/rules.mk .; fi || true
	//
	//	# authors
	//	if [ -f rules.mk ]; then make generate.authors; git add AUTHORS; fi || true
	//
	//	# copyright
	//	set -xe; \
	//	for prefix in "©" "Copyright" "Copyright (c)"; do \
	//		for file in README.md LICENSE-APACHE LICENSE-MIT LICENSE COPYRIGHT; do \
	//			if [ -f "$$file" ]; then \
	//				sed -i "s/$$prefix 2014 /$$prefix 2014-2021 /" $$file; \
	//				sed -i "s/$$prefix 2015 /$$prefix 2015-2021 /" $$file; \
	//				sed -i "s/$$prefix 2016 /$$prefix 2016-2021 /" $$file; \
	//				sed -i "s/$$prefix 2017 /$$prefix 2017-2021 /" $$file; \
	//				sed -i "s/$$prefix 2018 /$$prefix 2018-2021 /" $$file; \
	//				sed -i "s/$$prefix 2019 /$$prefix 2019-2021 /" $$file; \
	//				sed -i "s/$$prefix 2020 /$$prefix 2020-2021 /" $$file; \
	//				sed -i "s/$$prefix \([0-9][0-9][0-9][0-9]\)-20[0-9][0-9] /$$prefix \1-2021 /" $$file; \
	//				sed -i "s/$$prefix 2021-2021/$$prefix 2021 /" $$file; \
	//			fi; \
	//		done; \
	//	done
	//
	//	# golangci-lint fix
	//	sed -i "s/version: v1.26/version: v1.38/" .github/workflows/*.yml || true
	//	sed -i "s/version: v1.27/version: v1.38/" .github/workflows/*.yml || true
	//	sed -i "s/version: v1.28/version: v1.38/" .github/workflows/*.yml || true
	//
	//	# apply changes
	//	git diff
	//	git diff --cached
	//	git branch -D dev/moul/maintenance || true
	//	git checkout -b dev/moul/maintenance
	//	git status
	//	git commit -s -a -m "chore: repo maintenance 🤖" -m "more details: https://github.com/moul/repoman"
	//	git push -u origin dev/moul/maintenance -f
	//	hub pull-request -m "chore: repo maintenance 🤖" -m "more details: https://github.com/moul/repoman" || $(MAKE) -f $(REPOMAN)/Makefile _do.prlist

	return nil
}
