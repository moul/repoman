package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/mod/modfile"
	"moul.io/multipmuri"
	"moul.io/u"
)

type project struct {
	Path string
	Git  struct {
		Root          string
		MainBranch    string
		CurrentBranch string
		OriginRemotes []string
		InMainBranch  bool
		IsDirty       bool
		CloneURL      string
		HTMLURL       string
		RepoName      string
		RepoOwner     string
		Metadata      struct {
			HasGo      *bool  `json:"HasGo,omitempty"`
			HasDocker  *bool  `json:"HasDocker,omitempty"`
			HasLibrary *bool  `json:"HasLibrary,omitempty"`
			HasBinary  *bool  `json:"HasBinary,omitempty"`
			GoModPath  string `json:"GoModPath,omitempty"`
		} `json:"Metadata,omitempty"`

		head     *plumbing.Reference
		repo     *git.Repository
		origin   *git.Remote
		workTree *git.Worktree
		status   git.Status
	}
}

// nolint:nestif,gocognit
func projectFromPath(path string) (*project, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("incorrect path: %q: %w", path, err)
	}
	path = abs

	if !u.DirExists(path) {
		return nil, fmt.Errorf("path is not a directory: %q", path) // nolint:goerr113
	}

	project := &project{Path: path}
	project.Git.Root = gitFindRootDir(path)
	if project.Git.Root != "" {
		// open local repo
		{
			repo, err := git.PlainOpen(project.Git.Root)
			if err != nil {
				return nil, fmt.Errorf("failed to open git repo: %q: %w", project.Git.Root, err)
			}
			project.Git.repo = repo
		}

		// current branch
		{
			head, err := project.Git.repo.Head()
			if err != nil {
				return nil, fmt.Errorf("failed to get HEAD: %w", err)
			}
			project.Git.head = head
			project.Git.CurrentBranch = project.Git.head.Name().Short()
		}

		// 'origin' remote
		{
			origin, err := project.Git.repo.Remote("origin")
			if err != nil {
				return nil, fmt.Errorf("failed to get 'origin' remote: %w", err)
			}
			project.Git.origin = origin
			project.Git.OriginRemotes = origin.Config().URLs
			project.Git.CloneURL = origin.Config().URLs[0]
			ret, err := multipmuri.DecodeString(project.Git.CloneURL)
			if err != nil {
				return nil, fmt.Errorf("failed to parse the clone URL: %w", err)
			}
			repo := multipmuri.RepoEntity(ret)
			if typed, ok := repo.(*multipmuri.GitHubRepo); ok {
				project.Git.RepoName = typed.RepoID()
				project.Git.RepoOwner = typed.OwnerID()
				project.Git.HTMLURL = typed.String()
			} else {
				// FIXME: guess name
				// FIXME: guess owner
				panic("not implemented")
			}
		}

		// main branch
		{
			logger.Debug("rep.Reference(refs/remotes/origin/HEAD)")
			ref, err := project.Git.repo.Reference("refs/remotes/origin/HEAD", true)
			if err == nil {
				project.Git.MainBranch = strings.TrimPrefix(ref.Name().Short(), "origin/")
			} else { // if it fails, we try to fetch origin and then we retry
				logger.Debug("origin.List()")
				refs, err := project.Git.origin.List(&git.ListOptions{})
				if err != nil {
					logger.Warn("failed to list origin refs", zap.Error(err))
					project.Git.MainBranch = "n/a"
				}
				for _, ref := range refs {
					if ref.Name() == "HEAD" {
						project.Git.MainBranch = ref.Target().Short()
					}
				}
			}
		}
		if project.Git.MainBranch != "n/a" && project.Git.MainBranch != "" && project.Git.CurrentBranch != "" {
			project.Git.InMainBranch = project.Git.MainBranch == project.Git.CurrentBranch
		}

		// work tree
		{
			workTree, err := project.Git.repo.Worktree()
			if err != nil {
				return nil, fmt.Errorf("failed to get worktree: %w", err)
			}
			project.Git.workTree = workTree
		}

		// status
		logger.Debug("status")
		if err := project.updateStatus(); err != nil {
			return nil, fmt.Errorf("update status: %w", err)
		}

		// metadata
		{
			logger.Debug("guess metadata")
			// guess it
			if u.FileExists(filepath.Join(project.Path, "Dockerfile")) { // FIXME: look for other dockerfiles
				project.Git.Metadata.HasDocker = u.BoolPtr(true)
				project.Git.Metadata.HasBinary = u.BoolPtr(true)
			} else {
				project.Git.Metadata.HasDocker = u.BoolPtr(false)
				if u.FileExists(filepath.Join(project.Path, "main.go")) { // FIXME: improve check
					project.Git.Metadata.HasBinary = u.BoolPtr(true)
				} else {
					project.Git.Metadata.HasBinary = u.BoolPtr(false)
				}
			}
			if u.FileExists(filepath.Join(project.Path, "go.mod")) {
				project.Git.Metadata.HasGo = u.BoolPtr(true)
				content, err := ioutil.ReadFile(filepath.Join(project.Path, "go.mod"))
				if err != nil {
					return nil, fmt.Errorf("read go.mod: %w", err)
				}
				project.Git.Metadata.GoModPath = modfile.ModulePath(content)
			} else {
				goFiles, err := filepath.Glob(filepath.Join(project.Path, "*.go")) // FIXME: recursive
				if err != nil {
					return nil, fmt.Errorf("glob: %w", err)
				}
				project.Git.Metadata.HasGo = u.BoolPtr(len(goFiles) > 0)
			}
			project.Git.Metadata.HasLibrary = u.BoolPtr(*project.Git.Metadata.HasGo && !*project.Git.Metadata.HasBinary)

			// override it from metadata file
			// FIXME: TODO
		}
	} else {
		logger.Warn("project not within a git directory", zap.String("path", path))
	}

	return project, nil
}

func gitFindRootDir(path string) string {
	for {
		if u.DirExists(filepath.Join(path, ".git")) {
			return path
		}
		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}
	return ""
}

func (p *project) updateStatus() error {
	status, err := p.Git.workTree.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}
	p.Git.status = status
	p.Git.IsDirty = !status.IsClean()
	return nil
}

func (p *project) prepareWorkspace(opts projectOpts) error {
	if p.Git.Root == "" {
		return fmt.Errorf("not implemented: non-git projects")
	}

	if p.Git.IsDirty && opts.Reset {
		// reset
		{
			err := p.Git.workTree.Reset(&git.ResetOptions{
				Mode: git.HardReset,
			})
			if err != nil {
				return fmt.Errorf("reset worktree: %w", err)
			}
		}
		if err := p.updateStatus(); err != nil {
			return fmt.Errorf("update status: %w", err)
		}
	}
	if p.Git.IsDirty {
		return fmt.Errorf("worktree is dirty, please commit or discard changes before retrying") // nolint:goerr113
	}

	if opts.Fetch {
		logger.Debug("fetch origin", zap.String("project", p.Path))
		err := p.Git.origin.Fetch(&git.FetchOptions{
			Progress: os.Stderr,
		})
		switch err {
		case git.NoErrAlreadyUpToDate:
			// skip
		case nil:
			// skip
		default:
			return fmt.Errorf("failed to fetch origin: %w", err)
		}
	}

	if opts.CheckoutMainBranch && !p.Git.InMainBranch {
		logger.Debug("project is not using the main branch",
			zap.String("current", p.Git.CurrentBranch),
			zap.String("main", p.Git.MainBranch),
		)
		mainBranch, err := p.Git.repo.Branch(p.Git.MainBranch)
		if err != nil {
			return fmt.Errorf("failed to get ref for main branch: %q: %w", p.Git.MainBranch, err)
		}

		err = p.Git.workTree.Checkout(&git.CheckoutOptions{
			Branch: mainBranch.Merge,
		})
		if err != nil {
			return fmt.Errorf("failed to checkout main branch: %q: %w", p.Git.MainBranch, err)
		}

		err = p.Git.workTree.Pull(&git.PullOptions{})
		switch err {
		case git.NoErrAlreadyUpToDate: // skip
		case nil: // skip
		default:
			return fmt.Errorf("failed to pull main branch: %q: %w", p.Git.MainBranch, err)
		}
	}

	// check if the project looks like a one that can be maintained by repoman
	{
		var errs error
		for _, expected := range []string{"Makefile", "rules.mk"} {
			if !u.FileExists(filepath.Join(p.Path, expected)) {
				errs = multierr.Append(errs, fmt.Errorf("missing file: %q", expected))
			}
		}
		if errs != nil {
			return fmt.Errorf("project is not compatible with repoman: %w", errs)
		}
	}

	return nil
}

func (p *project) showDiff() error {
	script := `
		main() {
			# apply changes
			git diff
			git diff --cached
			git status
		}
		main
	`
	cmd := exec.Command("/bin/sh", "-xec", script)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Dir = p.Path
	cmd.Env = os.Environ()

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("publish script execution failed: %w", err)
	}
	return nil
}

func (p *project) openPR(branchName string, title string) error {
	logger.Debug("opening a PR", zap.String("branch", branchName), zap.String("title", title))
	initMoulBotEnv()
	script := `
		main() {
			# apply changes
			git branch -D {{.branchName}} || true
			git checkout -b {{.branchName}}
			git commit -s -a -m {{.title}} -m {{.body}}
			git push -u origin {{.branchName}} -f
			hub pull-request -m {{.title}} -m {{.body}} || hub pr list -f "- %pC%>(8)%i%Creset %U - %t% l%n"
		}
		main
	`
	body := "more details: https://github.com/moul/repoman"
	script = strings.ReplaceAll(script, "{{.branchName}}", fmt.Sprintf("%q", branchName))
	script = strings.ReplaceAll(script, "{{.title}}", fmt.Sprintf("%q", title))
	script = strings.ReplaceAll(script, "{{.body}}", fmt.Sprintf("%q", body))
	cmd := exec.Command("/bin/sh", "-xec", script)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Dir = p.Path
	cmd.Env = os.Environ()

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("publish script execution failed: %w", err)
	}
	return nil
}

func (p *project) pushChanges(opts projectOpts, branchName string, prTitle string) error {
	if opts.ShowDiff {
		err := p.showDiff()
		if err != nil {
			return fmt.Errorf("show diff: %w", err)
		}
	}

	if opts.OpenPR {
		err := p.openPR(branchName, prTitle)
		if err != nil {
			return fmt.Errorf("open PR: %w", err)
		}
	}
	return nil
}
