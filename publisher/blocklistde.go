package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/blocklistde"
)

const blocklistdeFile = "blocklistde.txt"

func fetchBlocklistde() ([]byte, error) {
	a := blocklistde.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncBlocklistdeData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(blocklistdeFile, data, wt, fs)
}
