package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/ohdear"
)

const ohdearFile = "ohdear.txt"

func fetchOhDear() ([]byte, error) {
	a := ohdear.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncOhDearData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(ohdearFile, data, wt, fs)
}
