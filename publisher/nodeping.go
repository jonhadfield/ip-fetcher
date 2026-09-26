package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/nodeping"
)

const nodepingFile = "nodeping.txt"

func fetchNodePing() ([]byte, error) {
	a := nodeping.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncNodePingData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(nodepingFile, data, wt, fs)
}
