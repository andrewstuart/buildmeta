package buildmeta

import (
	"encoding/json"
	"fmt"
	"go/build"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/src-d/go-git.v4"
)

const (
	noTagValue    = "NOT_TAGGED"
	buildmetaPath = "github.com/andrewstuart/buildmeta"
	noRemoteURL   = "NO_GIT_REMOTES_FOUND"
)

var (
	gitCommit      string
	gitCommitTime  string
	gitRepo        string
	gitTag         string
	buildTime      string
	goBuildVersion string
)

// This allows for debugging of binaries to see if the values were properly
// set, or to determine what versions etc a binary was built with.
func init() {
	if os.Getenv("BUILDMETA_TEST_DEBUG_AND_DIE") == "yes-i-want-my-app-to-exit-fatally-immediately" {
		bs, _ := json.MarshalIndent(GetInfo(), "", "  ")
		os.Stderr.Write(bs)
		os.Exit(1)
	}
}

// Info is the type used by the buildmeta package to return build-time
// information.
type Info struct {
	GitCommit      string `json:"commit"`
	GitCommitTime  string `json:"commitTime"`
	GitRepo        string `json:"gitRepo"`
	GitTag         string `json:"tag"`
	BuildTime      string `json:"buildTime"`
	GoBuildVersion string `json:"goBuildVersion,omitempty"`

	repoPath string
}

// PrometheusCollector returns a prometheus collector that can be registered by
// applications using buildmeta in order to provide standardized prometheus
// metrics on current version and builds.
func (i Info) PrometheusCollector() prometheus.Collector {
	g := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "buildmeta_info",
		Help: "Useful VCS/build info from the buildmeta library",
	}, []string{"commit", "commit_time", "tag", "build_time"})

	ct := i.GitCommitTime
	if t, err := time.Parse(time.RFC3339, ct); err == nil {
		ct = fmt.Sprint(t.Unix()) // For parsing by prometheus' time functions
	}
	bt := i.BuildTime
	if t, err := time.Parse(time.RFC3339, bt); err == nil {
		bt = fmt.Sprint(t.Unix()) // For parsing by prometheus' time functions
	}

	g.WithLabelValues(i.GitCommit, ct, i.GitTag, bt).Set(1)

	return g
}

func (i Info) TagOrCommit() string {
	if i.GitTag != noTagValue {
		return i.GitTag
	}
	return i.GitCommit
}

func (i Info) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(i)
}

// GetInfo returns the Info object with the values that were set by ldflags
// during compilation.
func GetInfo() Info {
	repo := gitRepo
	if u, err := url.Parse(repo); err == nil {
		u.User = nil // don't leak creds just in case
		repo = u.String()
	}

	return Info{
		GitCommit:      gitCommit,
		GitCommitTime:  gitCommitTime,
		GitRepo:        repo,
		GitTag:         gitTag,
		BuildTime:      buildTime,
		GoBuildVersion: goBuildVersion,
	}
}

func findGit(fromPath string) string {
	p, err := filepath.Abs(fromPath)
	if err != nil {
		return fromPath
	}

	for p != "/" {
		if s, err := os.Stat(path.Join(p, ".git")); err == nil && s.IsDir() {
			return p
		}
		p = path.Dir(p)
	}
	return fromPath
}

func (i Info) getPath() string {
	// Check to see if our library is vendored before returning a specific path,
	// since this will matter.
	vendorPath := path.Join(i.repoPath, "vendor", buildmetaPath)
	s, err := os.Stat(vendorPath)
	if err != nil || os.Getenv("GO111MODULE") != "on" || !s.IsDir() {
		return buildmetaPath
	}
	major, minor, err := getGoVersion()
	if err != nil || (major == 1 && minor > 12) {
		return buildmetaPath
	}
	rel, err := filepath.Rel(build.Default.GOPATH+"/src", vendorPath)
	if err != nil {
		return buildmetaPath
	}
	return rel
}

// LDFlags returns the ldflags that match the variables in the biuldmeta package.
func (i Info) LDFlags() string {
	p := i.getPath()

	return strings.Join([]string{
		fmt.Sprintf("-X %s.gitCommit=%s", p, i.GitCommit),
		fmt.Sprintf("-X %s.gitCommitTime=%s", p, i.GitCommitTime),
		fmt.Sprintf("-X %s.gitRepo=%s", p, i.GitRepo),
		fmt.Sprintf("-X %s.gitTag=%s", p, i.GitTag),
		fmt.Sprintf("-X %s.buildTime=%s", p, i.BuildTime),
		fmt.Sprintf("-X %s.goBuildVersion=%s", p, i.GoBuildVersion),
	}, " ")
}

// GenerateInfo returns the *Info object per a given repoPath on the local
// filesystem. It does not yet work outside a local filesystem.
func GenerateInfo(repoPath string) (*Info, error) {
	repoPath = findGit(repoPath)
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not open git repo")
	}

	head, err := repo.Head()
	if err != nil {
		return nil, errors.Wrap(err, "could not get current commit")
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, errors.Wrap(err, "could not load commit object for current head")
	}

	repoURL := noRemoteURL
	remotes, _ := repo.Remotes()
	if len(remotes) > 0 && len(remotes[0].Config().URLs) > 0 {
		repoURL = remotes[0].Config().URLs[0]
		if u, err := url.Parse(repoURL); err == nil {
			u.User = nil // don't leak creds just in case
			repoURL = u.String()
		}
	}

	tag := noTagValue
	tags, err := repo.Tags()
	if err == nil {
		for tt, err := tags.Next(); err == nil; tt, err = tags.Next() {
			if tt.Hash() == head.Hash() {
				tag = tt.Name().Short()
				break
			}
		}
		tags.Close()
	}

	work, err := repo.Worktree()
	if err != nil {
		return nil, errors.Wrap(err, "could not get working tree")
	}

	stat, err := work.Status()
	if err != nil {
		return nil, errors.Wrap(err, "could not get working tree status")
	}
	for k := range stat {
		if strings.HasSuffix(k, ".swo") || strings.HasSuffix(k, ".swp") {
			delete(stat, k)
		}
	}

	hString := head.Hash().String()
	if !stat.IsClean() && os.Getenv("CI") == "" {
		hString += "-dirty"
	}

	ver := runtime.Version()

	return &Info{
		GitCommit:      hString,
		GitCommitTime:  commit.Author.When.Format(time.RFC3339),
		GitRepo:        repoURL,
		GitTag:         tag,
		BuildTime:      time.Now().Format(time.RFC3339),
		GoBuildVersion: ver,
		repoPath:       repoPath,
	}, nil
}

func getGoVersion() (int, int, error) {
	out, err := exec.Command("go", "version").CombinedOutput()
	if err != nil {
		return 0, 0, fmt.Errorf("error running go: %w: %s", err, out)
	}

	fs := strings.Fields(string(out))
	if len(fs) < 3 {
		return 0, 0, fmt.Errorf("not enough fields returned by `go version`: %s", out)
	}

	var i, j int
	_, err = fmt.Sscanf(fs[2], "go%d.%d", &i, &j)
	if err != nil {
		return 0, 0, fmt.Errorf("error scanning version numbers: %w", err)
	}
	return i, j, nil
}
