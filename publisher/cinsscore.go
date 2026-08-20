package publisher

import (
	"bytes"
	"log/slog"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/cinsscore"
)

const cinsscoreFile = "cinsscore.txt"

func fetchCinsscore() ([]byte, error) {
	a := cinsscore.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncCinsscoreData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	rgb, err := fs.Open(cinsscoreFile)
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}

	if err == nil {
		upToDate, utdErr := isUpToDate(bytes.NewReader(data), rgb)
		if utdErr != nil || upToDate {
			return plumbing.ZeroHash, utdErr
		}

		slog.Info(cinsscoreFile, "up to date", upToDate)
	}

	if err = createFile(fs, cinsscoreFile, data); err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err = wt.Add(cinsscoreFile); err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update cinsscore data")
}
