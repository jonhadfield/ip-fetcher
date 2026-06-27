package publisher

import (
	"bytes"
	"log/slog"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/cdn77"
)

const cdn77File = "cdn77.json"

func fetchCDN77() ([]byte, error) {
	a := cdn77.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncCDN77Data(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	rgb, err := fs.Open(cdn77File)
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}

	if err == nil {
		upToDate, utdErr := isUpToDate(bytes.NewReader(data), rgb)
		if utdErr != nil || upToDate {
			return plumbing.ZeroHash, utdErr
		}

		slog.Info(cdn77File, "up to date", upToDate)
	}

	if err = createFile(fs, cdn77File, data); err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err = wt.Add(cdn77File); err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update cdn77 data")
}
