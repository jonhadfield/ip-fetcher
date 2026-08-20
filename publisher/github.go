package publisher

import (
	"bytes"
	"log/slog"
	"os"

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
	rgb, err := fs.Open(githubFile)
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}

	if err == nil {
		upToDate, utdErr := isUpToDate(bytes.NewReader(data), rgb)
		if utdErr != nil || upToDate {
			return plumbing.ZeroHash, utdErr
		}

		slog.Info(githubFile, "up to date", upToDate)
	}

	if err = createFile(fs, githubFile, data); err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err = wt.Add(githubFile); err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update github data")
}
