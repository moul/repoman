package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	hub "github.com/github/hub/v2/github"
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

	if opts.Maintenance.CheckoutMainBranch && !project.Git.InMainBranch {
		logger.Debug("project is not using the main branch",
			zap.String("current", project.Git.CurrentBranch),
			zap.String("main", project.Git.MainBranch),
		)
		mainBranch, err := project.Git.repo.Branch(project.Git.MainBranch)
		if err != nil {
			return fmt.Errorf("failed to get ref for main branch: %q: %w", project.Git.MainBranch, err)
		}

		err = project.Git.workTree.Checkout(&git.CheckoutOptions{
			Branch: mainBranch.Merge,
		})
		if err != nil {
			return fmt.Errorf("failed to checkout main branch: %q: %w", project.Git.MainBranch, err)
		}

		err = project.Git.workTree.Pull(&git.PullOptions{})
		switch err {
		case git.NoErrAlreadyUpToDate: // skip
		case nil: // skip
		default:
			return fmt.Errorf("failed to pull main branch: %q: %w", project.Git.MainBranch, err)
		}
	}

	// check if the project looks like a one that can be maintained by repoman
	{
		var errs error
		for _, expected := range []string{"Makefile", "rules.mk"} {
			if !u.FileExists(filepath.Join(project.Path, expected)) {
				errs = multierr.Append(errs, fmt.Errorf("missing file: %q", expected))
			}
		}
		if errs != nil {
			return fmt.Errorf("project is not compatible with repoman: %w", errs)
		}
	}

	initMoulBotEnv()

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

	if opts.Maintenance.ShowDiff {
		script := `
		main() {
			# apply changes
			git diff
			git diff --cached
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
			return fmt.Errorf("publish script execution failed: %w", err)
		}
	}

	// open a PR for the changes
	{
		script := `
		main() {
			# apply changes
			git branch -D dev/moul/maintenance || true
			git checkout -b dev/moul/maintenance
			git commit -s -a -m "chore: repo maintenance 🤖" -m "more details: https://github.com/moul/repoman"
			git push -u origin dev/moul/maintenance -f
			hub pull-request -m "chore: repo maintenance 🤖" -m "more details: https://github.com/moul/repoman" || hub pr list -f "- %pC%>(8)%i%Creset %U - %t% l%n"
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
			return fmt.Errorf("publish script execution failed: %w", err)
		}
	}

	return nil
}

func initMoulBotEnv() {
	if os.Getenv("REPOMAN_INITED") == "true" {
		return
	}
	os.Setenv("REPOMAN_INITED", "true")
	os.Setenv("GIT_AUTHOR_NAME", "moul-bot")
	os.Setenv("GIT_COMMITTER_NAME", "moul-bot")
	os.Setenv("GIT_AUTHOR_EMAIL", "bot@moul.io")
	os.Setenv("GIT_COMMITTER_EMAIL", "bot@moul.io")
	os.Setenv("HUB_CONFIG", filepath.Join(os.Getenv("HOME"), ".config", "hub-moul-bot"))
	config := hub.CurrentConfig()
	os.Setenv("GITHUB_TOKEN", config.Hosts[0].AccessToken)
}
