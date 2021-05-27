package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/hokaccha/go-prettyjson"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"moul.io/u"
)

func doDoctor(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}
	paths := u.UniqueStrings(args)
	logger.Debug("doDoctor", zap.Any("opts", opts), zap.Strings("project", paths))

	var errs error
	g, ctx := errgroup.WithContext(ctx)
	for _, path := range paths {
		path := path
		g.Go(func() error {
			err := doDoctorOnce(ctx, path)
			if err != nil {
				errs = multierr.Append(errs, fmt.Errorf("%q: %w", path, err))
			}
			return nil
		})
	}
	_ = g.Wait()
	return errs
}

func doDoctorOnce(_ context.Context, path string) error {
	project, err := projectFromPath(path)
	if err != nil {
		return fmt.Errorf("invalid project: %w", err)
	}
	s, err := prettyjson.Marshal(project)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}
	fmt.Println(string(s))
	return nil
}
