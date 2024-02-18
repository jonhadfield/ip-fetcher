package publisher

import (
	"bytes"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/googlebot"
	"log/slog"
	"os"
)

func syncGooglebot(wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	a := googlebot.New()

	originContent, _, _, err := a.FetchData()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	rgb, err := fs.Open("googlebot.json")
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}
	// if the file doesn't exist, we need to create it
	if err == nil {
		var upToDate bool

		upToDate, err = isUpToDate(bytes.NewReader(originContent), rgb)
		if err != nil || upToDate {
			return plumbing.ZeroHash, err
		}

		slog.Info("aws.json", "up to date", upToDate)
	}

	if err = createFile(fs, "googlebot.json", originContent); err != nil {
		return plumbing.ZeroHash, err
	}

	// Adds the new file to the staging area.
	_, err = wt.Add("googlebot.json")
	if err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update googlebot data")
}
