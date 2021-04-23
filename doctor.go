package main

import (
	"context"
	"fmt"

	"github.com/hokaccha/go-prettyjson"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"moul.io/u"
)

func doDoctor(ctx context.Context, args []string) error {
	paths := u.UniqueStrings(args)
	logger.Debug("doDoctor", zap.Any("opts", opts), zap.Strings("project", paths))

	g, ctx := errgroup.WithContext(ctx)
	for _, path := range paths {
		path := path
		g.Go(func() error { return doDoctorOnce(ctx, path) })
	}
	return g.Wait()
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
