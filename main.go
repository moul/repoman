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
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

type Opts struct {
	Path        string
	Maintenance struct{}
	Doctor      struct{}
	Version     struct{}
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
	// fs.StringVar(&opts.Path, "p", ".", "project's path")
}

func run(args []string) error {
	rand.Seed(srand.Fast())

	// setup flags
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
