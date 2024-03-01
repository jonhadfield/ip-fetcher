package publisher

import (
	"bytes"
	"encoding/json"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/linode"
	"log/slog"
	"os"
)

const linodeFileName = "linode.json"

func getLinodeJSON() ([]byte, error) {
	a := linode.New()

	data, err := a.Fetch()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(data, "", "  ")
}

func syncLinode(wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	linodeJSON, err := getLinodeJSON()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	rgb, err := fs.Open(linodeFileName)
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}
	// if the file doesn't exist, we need to create it
	if err == nil {
		var upToDate bool

		upToDate, err = isUpToDate(bytes.NewReader(linodeJSON), rgb)
		if err != nil || upToDate {
			return plumbing.ZeroHash, err
		}

		slog.Info(linodeFileName, "up to date", upToDate)
	}

	if err = createFile(fs, linodeFileName, linodeJSON); err != nil {
		return plumbing.ZeroHash, err
	}

	// Adds the new file to the staging area.
	_, err = wt.Add(linodeFileName)
	if err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update linode data")
}
