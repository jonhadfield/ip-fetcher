package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/x4bnet"
)

const x4bnetFile = "x4bnet.json"

func fetchX4BNet() ([]byte, error) {
	a := x4bnet.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncX4BNetData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(x4bnetFile, data, wt, fs)
}
