package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/ahrefs"
)

const ahrefsFile = "ahrefs.json"

func fetchAhrefs() ([]byte, error) {
	a := ahrefs.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncAhrefsData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(ahrefsFile, data, wt, fs)
}
