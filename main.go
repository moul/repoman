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
	"moul.io/zapconfig"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

var (
	rootFs = flag.NewFlagSet("<root>", flag.ExitOnError)

	logger *zap.Logger
)

func run(args []string) error {
	rand.Seed(srand.Fast())

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
			{Name: "doctor", Exec: doDoctor},
			{Name: "maintenance", Exec: doMaintenance},
			{Name: "version", Exec: doVersion},
		},
	}

	if err := root.ParseAndRun(context.Background(), args); err != nil {
		return err
	}

	return nil
}

func doDoctor(ctx context.Context, args []string) error {
	logger.Info("DOCTOR")
	return nil
}

func doMaintenance(ctx context.Context, args []string) error {
	logger.Info("MAINTENANCE")

	return nil
}

func doVersion(ctx context.Context, args []string) error {
	fmt.Print(motd.Default())
	fmt.Println("version: n/a")
	return nil
}
