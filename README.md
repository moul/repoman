# repoman

📋 repo manager: some scripts I run against my repos

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/moul.io/repoman)
[![License](https://img.shields.io/badge/license-Apache--2.0%20%2F%20MIT-%2397ca00.svg)](https://github.com/moul/repoman/blob/main/COPYRIGHT)
[![GitHub release](https://img.shields.io/github/release/moul/repoman.svg)](https://github.com/moul/repoman/releases)
[![Docker Metrics](https://images.microbadger.com/badges/image/moul/repoman.svg)](https://microbadger.com/images/moul/repoman)
[![Made by Manfred Touron](https://img.shields.io/badge/made%20by-Manfred%20Touron-blue.svg?style=flat)](https://manfred.life/)

[![Go](https://github.com/moul/repoman/workflows/Go/badge.svg)](https://github.com/moul/repoman/actions?query=workflow%3AGo)
[![Release](https://github.com/moul/repoman/workflows/Release/badge.svg)](https://github.com/moul/repoman/actions?query=workflow%3ARelease)
[![PR](https://github.com/moul/repoman/workflows/PR/badge.svg)](https://github.com/moul/repoman/actions?query=workflow%3APR)
[![GolangCI](https://golangci.com/badges/github.com/moul/repoman.svg)](https://golangci.com/r/github.com/moul/repoman)
[![codecov](https://codecov.io/gh/moul/repoman/branch/main/graph/badge.svg)](https://codecov.io/gh/moul/repoman)
[![Go Report Card](https://goreportcard.com/badge/moul.io/repoman)](https://goreportcard.com/report/moul.io/repoman)
[![CodeFactor](https://www.codefactor.io/repository/github/moul/repoman/badge)](https://www.codefactor.io/repository/github/moul/repoman)

[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/moul/repoman)

## Usage

[embedmd]:# (.tmp/example-info.txt console)
```console
foo@bar:~$ repoman info .
{
  "Git": {
    "CloneURL": "git@github.com:moul/repoman",
    "CurrentBranch": "master",
    "HTMLURL": "https://github.com/moul/repoman",
    "InMainBranch": true,
    "IsDirty": null,
    "MainBranch": "master",
    "Metadata": {
      "GoModPath": "moul.io/repoman",
      "HasBinary": true,
      "HasDocker": true,
      "HasGo": true,
      "HasLibrary": false
    },
    "OriginRemotes": [
      "git@github.com:moul/repoman"
    ],
    "RepoName": "repoman",
    "RepoOwner": "moul",
    "Root": "/home/moul/go/src/moul.io/repoman"
  },
  "Path": "/home/moul/go/src/moul.io/repoman"
}
```

---

[embedmd]:# (.tmp/usage.txt console)
```console
foo@bar:~$ repoman -h
USAGE
  repoman <subcommand>

SUBCOMMANDS
  info                 get project info
  doctor               perform various checks (read-only)
  maintenance          perform various maintenance tasks (write)
  version              show version and build info
  template-post-clone  replace template
  assets-config        generate a configuration for assets

FLAGS
  -v false  verbose mode
```

[embedmd]:# (.tmp/usage-maintenance.txt console)
```console
foo@bar:~$ repoman maintenance -h
USAGE
  maintenance [opts] <path...>

FLAGS
  -bump-deps false            bump dependencies
  -checkout-main-branch true  switch to the main branch before applying the changes
  -fetch true                 fetch origin before applying the changes
  -open-pr true               open a new pull-request with the changes
  -reset false                reset dirty worktree before applying the changes
  -show-diff true             display git diff of the changes
  -std true                   standard maintenance tasks
```

[embedmd]:# (.tmp/usage-info.txt console)
```console
foo@bar:~$ repoman info -h
USAGE
  info [opts] <path...>
```

[embedmd]:# (.tmp/usage-template-post-clone.txt console)
```console
foo@bar:~$ repoman template-post-clone -h
USAGE
  template-post-clone [opts] <path...>

FLAGS
  -checkout-main-branch true           switch to the main branch before applying the changes
  -fetch true                          fetch origin before applying the changes
  -open-pr true                        open a new pull-request with the changes
  -reset false                         reset dirty worktree before applying the changes
  -rm-go-binary false                  whether to delete everything related to go binary and only keep a library
  -show-diff true                      display git diff of the changes
  -template-name golang-repo-template  template's name (to change with the new project's name)
  -template-owner moul                 template owner's name (to change with the new owner)
```

## Install

### Using go

```sh
go get moul.io/repoman
```

### Releases

See https://github.com/moul/repoman/releases

## Contribute

![Contribute <3](https://raw.githubusercontent.com/moul/moul/main/contribute.gif)

I really welcome contributions.
Your input is the most precious material.
I'm well aware of that and I thank you in advance.
Everyone is encouraged to look at what they can do on their own scale;
no effort is too small.

Everything on contribution is sum up here: [CONTRIBUTING.md](./CONTRIBUTING.md)

### Dev helpers

Pre-commit script for install: https://pre-commit.com

### Contributors ✨

<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-2-orange.svg)](#contributors)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="http://manfred.life"><img src="https://avatars1.githubusercontent.com/u/94029?v=4" width="100px;" alt=""/><br /><sub><b>Manfred Touron</b></sub></a><br /><a href="#maintenance-moul" title="Maintenance">🚧</a> <a href="https://github.com/moul/repoman/commits?author=moul" title="Documentation">📖</a> <a href="https://github.com/moul/repoman/commits?author=moul" title="Tests">⚠️</a> <a href="https://github.com/moul/repoman/commits?author=moul" title="Code">💻</a></td>
    <td align="center"><a href="https://manfred.life/moul-bot"><img src="https://avatars1.githubusercontent.com/u/41326314?v=4" width="100px;" alt=""/><br /><sub><b>moul-bot</b></sub></a><br /><a href="#maintenance-moul-bot" title="Maintenance">🚧</a></td>
  </tr>
</table>

<!-- markdownlint-enable -->
<!-- prettier-ignore-end -->
<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors)
specification. Contributions of any kind welcome!

### Stargazers over time

[![Stargazers over time](https://starchart.cc/moul/repoman.svg)](https://starchart.cc/moul/repoman)

## License

© 2021   [Manfred Touron](https://manfred.life)

Licensed under the [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)
([`LICENSE-APACHE`](LICENSE-APACHE)) or the [MIT license](https://opensource.org/licenses/MIT)
([`LICENSE-MIT`](LICENSE-MIT)), at your option.
See the [`COPYRIGHT`](COPYRIGHT) file for more details.

`SPDX-License-Identifier: (Apache-2.0 OR MIT)`
