package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/contabo"
)

const contaboFile = "contabo.json"

func fetchContabo() ([]byte, error) {
	a := contabo.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncContaboData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(contaboFile, data, wt, fs)
}
