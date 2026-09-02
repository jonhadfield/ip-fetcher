package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/tenable"
)

const tenableFile = "tenable.json"

func fetchTenable() ([]byte, error) {
	a := tenable.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncTenableData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(tenableFile, data, wt, fs)
}
