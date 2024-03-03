package publisher

import (
	"bytes"
	"log/slog"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/googlesc"
)

const googlescFile = "googlesc.json"

func syncGoogleSC(wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	a := googlesc.New()

	originContent, _, _, err := a.FetchData()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	rgb, err := fs.Open(googlescFile)
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

		slog.Info(googlescFile, "up to date", upToDate)
	}

	if err = createFile(fs, googlescFile, originContent); err != nil {
		return plumbing.ZeroHash, err
	}

	// Adds the new file to the staging area.
	_, err = wt.Add(googlescFile)
	if err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update googlesc data")
}
