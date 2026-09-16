package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/cachefly"
)

const cacheflyFile = "cachefly.txt"

func fetchCacheFly() ([]byte, error) {
	a := cachefly.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncCacheFlyData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(cacheflyFile, data, wt, fs)
}
