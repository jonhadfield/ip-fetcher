package publisher

import (
	"bytes"
	"log/slog"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/icloudpr"
)

const icloudprFile = "icloudpr.csv"

func fetchICloudPR() ([]byte, error) {
	a := icloudpr.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncICloudPRData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	rgb, err := fs.Open(icloudprFile)
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}

	if err == nil {
		upToDate, utdErr := isUpToDate(bytes.NewReader(data), rgb)
		if utdErr != nil || upToDate {
			return plumbing.ZeroHash, utdErr
		}

		slog.Info(icloudprFile, "up to date", upToDate)
	}

	if err = createFile(fs, icloudprFile, data); err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err = wt.Add(icloudprFile); err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update icloudpr data")
}
