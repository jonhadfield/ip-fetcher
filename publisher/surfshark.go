package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/surfshark"
)

const surfsharkFile = "surfshark.json"

func fetchSurfshark() ([]byte, error) {
	a := surfshark.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncSurfsharkData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(surfsharkFile, data, wt, fs)
}
