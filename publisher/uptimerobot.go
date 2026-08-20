package publisher

import (
	"bytes"
	"log/slog"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/uptimerobot"
)

const uptimerobotFile = "uptimerobot.txt"

func fetchUptimerobot() ([]byte, error) {
	a := uptimerobot.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncUptimerobotData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	rgb, err := fs.Open(uptimerobotFile)
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}

	if err == nil {
		upToDate, utdErr := isUpToDate(bytes.NewReader(data), rgb)
		if utdErr != nil || upToDate {
			return plumbing.ZeroHash, utdErr
		}

		slog.Info(uptimerobotFile, "up to date", upToDate)
	}

	if err = createFile(fs, uptimerobotFile, data); err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err = wt.Add(uptimerobotFile); err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update uptimerobot data")
}
