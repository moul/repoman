package main

import (
	"context"
	"flag"
	"fmt"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/google/go-github/v35/github"
	"github.com/hokaccha/go-prettyjson"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"moul.io/u"
)

func doAssetsConfig(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}
	paths := u.UniqueStrings(args)
	logger.Debug("doAssetsConfig", zap.Any("opts", opts), zap.Strings("project", paths))

	var errs error
	g, ctx := errgroup.WithContext(ctx)
	for _, path := range paths {
		path := path
		g.Go(func() error {
			err := doAssetsConfigOnce(ctx, path)
			if err != nil {
				errs = multierr.Append(errs, fmt.Errorf("%q: %w", path, err))
			}
			return nil
		})
	}
	_ = g.Wait()
	return errs
}

type assetConfigVersion struct {
	TargetVersion string `json:",omitempty"`
	Assets        int    `json:",omitempty"`
}

type assetConfig struct {
	VersionAliases map[string]assetConfigVersion
	SemverMapping  map[string]string `json:",omitempty"`
}

func doAssetsConfigOnce(_ context.Context, path string) error {
	project, err := projectFromPath(path)
	if err != nil {
		return fmt.Errorf("invalid project: %w", err)
	}

	// fetch releases
	var releases []*github.RepositoryRelease
	{
		client := github.NewClient(nil)
		var err error
		releases, _, err = client.Repositories.ListReleases(context.Background(), project.Git.RepoOwner, project.Git.RepoName, nil)
		if err != nil {
			return fmt.Errorf("GH API: list releases: %w", err)
		}
	}

	// compute config
	var config assetConfig
	{
		config = assetConfig{
			VersionAliases: make(map[string]assetConfigVersion),
			SemverMapping:  make(map[string]string),
		}

		rawVersions := []string{}
		for _, release := range releases {
			if release.GetDraft() || release.GetPrerelease() || len(release.Assets) == 0 {
				logger.Debug("ignoring release",
					zap.String("version", release.GetTagName()),
					zap.Bool("draft", release.GetDraft()),
					zap.Bool("prerelease", release.GetPrerelease()),
					zap.Int("assets", len(release.Assets)),
				)
				continue
			}
			rawVersions = append(rawVersions, release.GetTagName())
		}
		versions := make([]*semver.Version, 0)
		for _, raw := range rawVersions {
			version, err := semver.NewVersion(raw)
			if err != nil {
				logger.Warn("cannot parse version", zap.String("raw", raw), zap.Error(err))
				continue
			}
			config.SemverMapping[version.String()] = raw
			versions = append(versions, version)
		}
		sort.Sort(semver.Collection(versions))

		for _, version := range versions {
			raw := config.SemverMapping[version.String()]
			var release *github.RepositoryRelease
			for _, r := range releases {
				if r.GetTagName() == raw {
					release = r
					break
				}
			}
			_ = release
			minor := fmt.Sprintf("v%d.%d", version.Major(), version.Minor())
			major := fmt.Sprintf("v%d", version.Major())
			configVersion := assetConfigVersion{
				TargetVersion: raw,
				Assets:        len(release.Assets),
			}
			config.VersionAliases[minor] = configVersion
			config.VersionAliases[major] = configVersion
			config.VersionAliases["latest"] = configVersion
		}
	}

	// cleanup
	{
		config.SemverMapping = nil
	}

	// print
	{
		s, err := prettyjson.Marshal(config)
		if err != nil {
			return fmt.Errorf("json marshal error: %w", err)
		}
		fmt.Println(string(s))
	}
	return nil
}
