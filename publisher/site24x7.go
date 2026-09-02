package publisher

import (
	"bytes"
	"log/slog"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/site24x7"
)

const site24x7File = "site24x7.json"

func fetchSite24x7() ([]byte, error) {
	a := site24x7.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncSite24x7Data(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	rgb, err := fs.Open(site24x7File)
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}

	if err == nil {
		upToDate, utdErr := isUpToDate(bytes.NewReader(data), rgb)
		if utdErr != nil || upToDate {
			return plumbing.ZeroHash, utdErr
		}

		slog.Info(site24x7File, "up to date", upToDate)
	}

	if err = createFile(fs, site24x7File, data); err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err = wt.Add(site24x7File); err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update site24x7 data")
}
