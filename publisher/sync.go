package publisher

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// syncData writes a fetched document to name when it differs from the copy the
// repository already holds, staging and committing the change. The zero hash is
// returned when the file is already up to date.
func syncData(name string, data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	existing, err := fs.Open(name)
	if err != nil && !os.IsNotExist(err) {
		return plumbing.ZeroHash, err
	}

	if err == nil {
		upToDate, utdErr := isUpToDate(bytes.NewReader(data), existing)
		if utdErr != nil || upToDate {
			return plumbing.ZeroHash, utdErr
		}

		slog.Info(name, "up to date", upToDate)
	}

	if err = createFile(fs, name, data); err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err = wt.Add(name); err != nil {
		return plumbing.ZeroHash, err
	}

	return createCommit(wt, "update "+strings.TrimSuffix(name, filepath.Ext(name))+" data")
}
