package publisher

import (
	"bytes"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/oci"
	"log/slog"
	"os"
)

const ociFile = "oci.json"

func syncOCI(wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	a := oci.New()

	originContent, _, _, err := a.FetchData()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	rgb, err := fs.Open(ociFile)
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

		slog.Info(ociFile, "up to date", upToDate)
	}

	if err = createFile(fs, ociFile, originContent); err != nil {
		return plumbing.ZeroHash, err
	}

	// Adds the new file to the staging area.
	_, err = wt.Add(ociFile)
	if err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update oci data")
}
