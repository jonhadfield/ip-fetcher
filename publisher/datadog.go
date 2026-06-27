package publisher

import (
	"bytes"
	"log/slog"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/datadog"
)

const datadogFile = "datadog.json"

func fetchDatadog() ([]byte, error) {
	a := datadog.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncDatadogData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	rgb, err := fs.Open(datadogFile)
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}

	if err == nil {
		upToDate, utdErr := isUpToDate(bytes.NewReader(data), rgb)
		if utdErr != nil || upToDate {
			return plumbing.ZeroHash, utdErr
		}

		slog.Info(datadogFile, "up to date", upToDate)
	}

	if err = createFile(fs, datadogFile, data); err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err = wt.Add(datadogFile); err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update datadog data")
}
