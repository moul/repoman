package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"moul.io/srand"
	"moul.io/zapconfig"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		if err != flag.ErrHelp {
			fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		}
		os.Exit(1)
	}
}

type projectOpts struct {
	CheckoutMainBranch bool
	Fetch              bool
	ShowDiff           bool
	OpenPR             bool
	Reset              bool
}

type Opts struct {
	Verbose     bool
	Path        string
	Maintenance struct {
		Project  projectOpts
		BumpDeps bool
		Standard bool
	}
	TemplatePostClone struct {
		Project        projectOpts
		RemoveGoBinary bool
		TemplateName   string
		TemplateOwner  string
	}
	Doctor  struct{}
	Info    struct{}
	Version struct{}
}

var (
	rootFs              = flag.NewFlagSet("<root>", flag.ExitOnError)
	doctorFs            = flag.NewFlagSet("doctor", flag.ExitOnError)
	infoFs              = flag.NewFlagSet("doctor", flag.ExitOnError)
	maintenanceFs       = flag.NewFlagSet("maintenance", flag.ExitOnError)
	versionFs           = flag.NewFlagSet("version", flag.ExitOnError)
	templatePostCloneFs = flag.NewFlagSet("template-post-clone", flag.ExitOnError)
	assetsConfigFs      = flag.NewFlagSet("assets-config", flag.ExitOnError)
	opts                Opts

	logger *zap.Logger
)

func run(args []string) error {
	rand.Seed(srand.Fast())

	// setup flags
	{
		setupProjectFlags := func(fs *flag.FlagSet, opts *projectOpts) {
			fs.BoolVar(&opts.CheckoutMainBranch, "checkout-main-branch", true, "switch to the main branch before applying the changes")
			fs.BoolVar(&opts.Fetch, "fetch", true, "fetch origin before applying the changes")
			fs.BoolVar(&opts.ShowDiff, "show-diff", true, "display git diff of the changes")
			fs.BoolVar(&opts.OpenPR, "open-pr", true, "open a new pull-request with the changes")
			fs.BoolVar(&opts.Reset, "reset", false, "reset dirty worktree before applying the changes")
		}
		rootFs.BoolVar(&opts.Verbose, "v", false, "verbose mode")
		setupProjectFlags(templatePostCloneFs, &opts.TemplatePostClone.Project)
		templatePostCloneFs.StringVar(&opts.TemplatePostClone.TemplateName, "template-name", "golang-repo-template", "template's name (to change with the new project's name)")
		templatePostCloneFs.StringVar(&opts.TemplatePostClone.TemplateOwner, "template-owner", "moul", "template owner's name (to change with the new owner)")
		templatePostCloneFs.BoolVar(&opts.TemplatePostClone.RemoveGoBinary, "rm-go-binary", false, "whether to delete everything related to go binary and only keep a library")
		setupProjectFlags(maintenanceFs, &opts.Maintenance.Project)
		maintenanceFs.BoolVar(&opts.Maintenance.BumpDeps, "bump-deps", false, "bump dependencies")
		maintenanceFs.BoolVar(&opts.Maintenance.Standard, "std", true, "standard maintenance tasks")
	}

	root := &ffcli.Command{
		Name:       "repoman",
		FlagSet:    rootFs,
		ShortUsage: "repoman <subcommand>",
		Subcommands: []*ffcli.Command{
			{Name: "info", Exec: doInfo, FlagSet: infoFs, ShortHelp: "get project info", ShortUsage: "info [opts] <path...>"},
			{Name: "doctor", Exec: doDoctor, FlagSet: doctorFs, ShortHelp: "perform various checks (read-only)", ShortUsage: "doctor [opts] <path...>"},
			{Name: "maintenance", Exec: doMaintenance, FlagSet: maintenanceFs, ShortHelp: "perform various maintenance tasks (write)", ShortUsage: "maintenance [opts] <path...>"},
			{Name: "version", Exec: doVersion, FlagSet: versionFs, ShortHelp: "show version and build info", ShortUsage: "version"},
			{Name: "template-post-clone", Exec: doTemplatePostClone, FlagSet: templatePostCloneFs, ShortHelp: "replace template", ShortUsage: "template-post-clone [opts] <path...>"},
			{Name: "assets-config", Exec: doAssetsConfig, FlagSet: assetsConfigFs, ShortHelp: "generate a configuration for assets", ShortUsage: "assets-config [opts] <path...>"},
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}

	if err := root.Parse(args); err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// init logger
	{
		var err error
		config := zapconfig.New().
			SetPreset("light-console")
		if opts.Verbose {
			config = config.SetLevel(zapcore.DebugLevel)
		} else {
			config = config.SetLevel(zapcore.InfoLevel)
		}
		logger, err = config.Build()
		if err != nil {
			return err
		}
	}

	if err := root.Run(context.Background()); err != nil {
		return fmt.Errorf("run error: %w", err)
	}
	return nil
}
