package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/flyio"
)

const flyioFile = "flyio.json"

func fetchFlyio() ([]byte, error) {
	a := flyio.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncFlyioData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(flyioFile, data, wt, fs)
}
