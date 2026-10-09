package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/bunny"
)

const bunnyFile = "bunny.json"

func fetchBunny() ([]byte, error) {
	a := bunny.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncBunnyData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(bunnyFile, data, wt, fs)
}
