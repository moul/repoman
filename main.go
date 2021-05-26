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

type Opts struct {
	Path        string
	Maintenance struct {
		CheckoutMainBranch bool
		NoFetch            bool
		BumpDeps           bool
		Standard           bool
		ShowDiff           bool
	}
	Doctor  struct{}
	Version struct{}
}

var (
	rootFs        = flag.NewFlagSet("<root>", flag.ExitOnError)
	doctorFs      = flag.NewFlagSet("doctor", flag.ExitOnError)
	maintenanceFs = flag.NewFlagSet("maintenance", flag.ExitOnError)
	versionFs     = flag.NewFlagSet("version", flag.ExitOnError)
	bumpDepsFs    = flag.NewFlagSet("bump-deps", flag.ExitOnError)
	opts          Opts

	logger *zap.Logger
)

func setupRootFlags(fs *flag.FlagSet) {
	// fs.StringVar(&opts.Path, "p", ".", "project's path")
}

func run(args []string) error {
	rand.Seed(srand.Fast())

	// setup flags
	setupRootFlags(rootFs)
	setupRootFlags(doctorFs)
	setupRootFlags(maintenanceFs)
	setupRootFlags(bumpDepsFs)
	maintenanceFs.BoolVar(&opts.Maintenance.CheckoutMainBranch, "checkout-main-branch", true, "switch to the main branch before applying the maintenance")
	maintenanceFs.BoolVar(&opts.Maintenance.NoFetch, "no-fetch", false, "do not fetch origin")
	maintenanceFs.BoolVar(&opts.Maintenance.BumpDeps, "bump-deps", false, "bump dependencies")
	maintenanceFs.BoolVar(&opts.Maintenance.Standard, "std", true, "standard maintenance tasks")
	maintenanceFs.BoolVar(&opts.Maintenance.ShowDiff, "show-diff", true, "display git diff of the changes")

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
			// copyTemplate
			// postTemplateClone
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
