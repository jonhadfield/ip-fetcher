package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/updown"
)

const updownFile = "updown.json"

func fetchUpdown() ([]byte, error) {
	a := updown.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncUpdownData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(updownFile, data, wt, fs)
}
