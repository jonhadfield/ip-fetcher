package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/internal/iplist"
	"github.com/jonhadfield/ip-fetcher/providers/github"
)

const githubFile = "github.txt"

// fetchGitHub returns the parsed prefixes rather than the raw response, as the
// meta endpoint also carries key fingerprints and feature flags.
func fetchGitHub() ([]byte, error) {
	gh := github.New()

	prefixes, err := gh.Fetch()
	if err != nil {
		return nil, err
	}

	return iplist.ToLines(prefixes), nil
}

func syncGitHubData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(githubFile, data, wt, fs)
}
