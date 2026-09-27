package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/xpanse"
)

const xpanseFile = "xpanse.txt"

func fetchXpanse() ([]byte, error) {
	a := xpanse.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncXpanseData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(xpanseFile, data, wt, fs)
}
