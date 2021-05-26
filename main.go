package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
	"go.uber.org/zap"
	"moul.io/srand"
	"moul.io/zapconfig"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		os.Exit(1)
	}
}

type projectOpts struct {
	CheckoutMainBranch bool
	Fetch              bool
	ShowDiff           bool
	OpenPR             bool
}

type Opts struct {
	Path        string
	Maintenance struct {
		Project  projectOpts
		BumpDeps bool
		Standard bool
	}
	TemplatePostClone struct {
		Project       projectOpts
		TemplateName  string
		TemplateOwner string
	}
	Doctor  struct{}
	Version struct{}
}

var (
	rootFs              = flag.NewFlagSet("<root>", flag.ExitOnError)
	doctorFs            = flag.NewFlagSet("doctor", flag.ExitOnError)
	maintenanceFs       = flag.NewFlagSet("maintenance", flag.ExitOnError)
	versionFs           = flag.NewFlagSet("version", flag.ExitOnError)
	templatePostCloneFs = flag.NewFlagSet("tmeplate-post-clone", flag.ExitOnError)
	opts                Opts

	logger *zap.Logger
)

func run(args []string) error {
	rand.Seed(srand.Fast())

	// setup flags
	{
		setupRootFlags := func(fs *flag.FlagSet) {
			// fs.StringVar(&opts.Path, "p", ".", "project's path")
		}
		setupProjectFlags := func(fs *flag.FlagSet, opts *projectOpts) {
			fs.BoolVar(&opts.CheckoutMainBranch, "checkout-main-branch", true, "switch to the main branch before applying the changes")
			fs.BoolVar(&opts.Fetch, "fetch", true, "fetch origin before applying the changes")
			fs.BoolVar(&opts.ShowDiff, "show-diff", true, "display git diff of the changes")
			fs.BoolVar(&opts.OpenPR, "open-pr", true, "open a new pull-request with the changes")
		}
		// root
		setupRootFlags(rootFs)

		// doctor
		setupRootFlags(doctorFs)

		// template-post-clone
		setupRootFlags(templatePostCloneFs)
		setupProjectFlags(templatePostCloneFs, &opts.TemplatePostClone.Project)
		templatePostCloneFs.StringVar(&opts.TemplatePostClone.TemplateName, "template-name", "golang-repo-template", "template's name (to change with the new project's name)")
		templatePostCloneFs.StringVar(&opts.TemplatePostClone.TemplateOwner, "template-owner", "moul", "template owner's name (to change with the new owner)")
		// moul.io/<name> -> github.com/something/<name>

		// version
		// n/a

		// maintenance
		setupRootFlags(maintenanceFs)
		setupProjectFlags(maintenanceFs, &opts.Maintenance.Project)
		maintenanceFs.BoolVar(&opts.Maintenance.BumpDeps, "bump-deps", false, "bump dependencies")
		maintenanceFs.BoolVar(&opts.Maintenance.Standard, "std", true, "standard maintenance tasks")
	}

	// init logger
	{
		var err error
		logger, err = (&zapconfig.Configurator{}).SetPreset("light-console").Build()
		if err != nil {
			return err
		}
	}

	root := &ffcli.Command{
		Name:    "repoman",
		FlagSet: rootFs,
		Subcommands: []*ffcli.Command{
			{Name: "doctor", Exec: doDoctor, FlagSet: doctorFs, ShortHelp: "perform various checks (read-only)"},
			{Name: "maintenance", Exec: doMaintenance, FlagSet: maintenanceFs, ShortHelp: "perform various maintenance tasks (write)"},
			{Name: "version", Exec: doVersion, FlagSet: versionFs, ShortHelp: "show version and build info"},
			{Name: "template-post-clone", Exec: doTemplatePostClone, FlagSet: templatePostCloneFs, ShortHelp: "replace template"},
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}

	if err := root.ParseAndRun(context.Background(), args); err != nil {
		return err
	}

	return nil
}
