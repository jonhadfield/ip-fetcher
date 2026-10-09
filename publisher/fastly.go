package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/fastly"
)

const fastlyFile = "fastly.json"

func fetchFastly() ([]byte, error) {
	a := fastly.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncFastlyData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(fastlyFile, data, wt, fs)
}
