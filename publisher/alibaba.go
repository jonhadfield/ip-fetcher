package publisher

import (
	"bytes"
	"log/slog"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/alibaba"
)

const alibabaFile = "alibaba.json"

func fetchAlibaba() ([]byte, error) {
	a := alibaba.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncAlibabaData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	rgb, err := fs.Open(alibabaFile)
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}

	if err == nil {
		upToDate, utdErr := isUpToDate(bytes.NewReader(data), rgb)
		if utdErr != nil || upToDate {
			return plumbing.ZeroHash, utdErr
		}

		slog.Info(alibabaFile, "up to date", upToDate)
	}

	if err = createFile(fs, alibabaFile, data); err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err = wt.Add(alibabaFile); err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update alibaba data")
}
