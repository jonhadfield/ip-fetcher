package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/cymru"
)

const cymruFile = "cymru.json"

func fetchCymru() ([]byte, error) {
	a := cymru.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncCymruData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(cymruFile, data, wt, fs)
}
