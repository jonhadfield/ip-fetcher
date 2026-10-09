package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/imperva"
)

const impervaFile = "imperva.json"

func fetchImperva() ([]byte, error) {
	a := imperva.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncImpervaData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(impervaFile, data, wt, fs)
}
