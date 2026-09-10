package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/feodo"
)

const feodoFile = "feodo.txt"

func fetchFeodo() ([]byte, error) {
	a := feodo.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncFeodoData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(feodoFile, data, wt, fs)
}
