package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
	"go.uber.org/zap"
	"moul.io/motd"
	"moul.io/srand"
	"moul.io/u"
	"moul.io/zapconfig"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

type Opts struct {
	Path string
}

var (
	rootFs        = flag.NewFlagSet("<root>", flag.ExitOnError)
	doctorFs      = flag.NewFlagSet("doctor", flag.ExitOnError)
	maintenanceFs = flag.NewFlagSet("maintenance", flag.ExitOnError)
	versionFs     = flag.NewFlagSet("version", flag.ExitOnError)
	opts          Opts

	logger *zap.Logger
)

func setupRootFlags(fs *flag.FlagSet) {
	fs.StringVar(&opts.Path, "p", ".", "project's path")
}

func run(args []string) error {
	rand.Seed(srand.Fast())

	setupRootFlags(rootFs)
	setupRootFlags(doctorFs)
	setupRootFlags(maintenanceFs)

	// init logger
	{
		var err error
		logger, err = zapconfig.Configurator{}.Build()
		if err != nil {
			return err
		}
	}

	root := &ffcli.Command{
		FlagSet: rootFs,
		Subcommands: []*ffcli.Command{
			{Name: "doctor", Exec: doDoctor, FlagSet: doctorFs},
			{Name: "maintenance", Exec: doMaintenance, FlagSet: maintenanceFs},
			{Name: "version", Exec: doVersion, FlagSet: versionFs},
			// copyTemplate
			// postTemplateClone
		},
	}

	if err := root.ParseAndRun(context.Background(), args); err != nil {
		return err
	}

	return nil
}

func doDoctor(ctx context.Context, args []string) error {
	project, err := projectFromPath(opts.Path)
	if err != nil {
		return fmt.Errorf("invalid project: %w", err)
	}
	logger.Debug("doDoctor", zap.Any("opts", opts), zap.Any("project", project))
	fmt.Println("## Project")
	fmt.Println(u.PrettyJSON(project))
	return nil
}

func doMaintenance(ctx context.Context, args []string) error {
	project, err := projectFromPath(opts.Path)
	if err != nil {
		return fmt.Errorf("invalid project: %w", err)
	}
	logger.Debug("doMaintenance", zap.Any("opts", opts), zap.Any("project", project))

	// - repoman.yml ->
	//   - template -> moul/golang-repo-template
	//   - exclude: - README.md
	//   - no-main / lib-only
	// - auto update from template

	// COMMANDS = hubsync checkoutmaster maintenance prlist
	// REPOS ?= $(wildcard */)
	// OPTS ?= ;
	// REPOMAN ?= ~/go/src/moul.io/repoman
	//
	// .PHONY: $(COMMANDS)
	// $(COMMANDS):
	//	@for repo in $(REPOS); do ( set -e; \
	//		echo "cd $$repo && make -s -f $(REPOMAN)/Makefile _do.$@ $(OPTS)"; \
	//		cd $$repo && make -s -f $(REPOMAN)/Makefile _do.$@ $(OPTS) \
	//	); done
	//
	// _do.checkoutmaster: _do.hubsync
	//	git checkout master
	//
	// _do.hubsync:
	//	hub sync
	//
	// _do.prlist:
	//	@hub pr list -f "- %pC%>(8)%i%Creset %U - %t% l%n"
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

func doVersion(ctx context.Context, args []string) error {
	fmt.Print(motd.Default())
	fmt.Println("version: n/a")
	return nil
}
