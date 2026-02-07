package publisher

import (
	"bytes"
	"log/slog"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/googleutf"
)

const googleutfFile = "googleutf.json"

func fetchGoogleUTF() ([]byte, error) {
	a := googleutf.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncGoogleUTFData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	rgb, err := fs.Open(googleutfFile)
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}

	if err == nil {
		upToDate, utdErr := isUpToDate(bytes.NewReader(data), rgb)
		if utdErr != nil || upToDate {
			return plumbing.ZeroHash, utdErr
		}

		slog.Info(googleutfFile, "up to date", upToDate)
	}

	if err = createFile(fs, googleutfFile, data); err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err = wt.Add(googleutfFile); err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update googleutf data")
}
