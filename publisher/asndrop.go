package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/asndrop"
)

const asndropFile = "asndrop.json"

func fetchASNDrop() ([]byte, error) {
	a := asndrop.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncASNDropData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(asndropFile, data, wt, fs)
}
