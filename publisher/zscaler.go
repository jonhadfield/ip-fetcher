package publisher

import (
	"bytes"
	"log/slog"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/zscaler"
)

const zscalerFile = "zscaler.json"

func fetchZscaler() ([]byte, error) {
	a := zscaler.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncZscalerData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	rgb, err := fs.Open(zscalerFile)
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}

	if err == nil {
		upToDate, utdErr := isUpToDate(bytes.NewReader(data), rgb)
		if utdErr != nil || upToDate {
			return plumbing.ZeroHash, utdErr
		}

		slog.Info(zscalerFile, "up to date", upToDate)
	}

	if err = createFile(fs, zscalerFile, data); err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err = wt.Add(zscalerFile); err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update zscaler data")
}
