package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/gcore"
)

const gcoreFile = "gcore.json"

func fetchGcore() ([]byte, error) {
	a := gcore.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncGcoreData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(gcoreFile, data, wt, fs)
}
