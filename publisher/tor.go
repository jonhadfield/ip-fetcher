package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/tor"
)

const torFile = "tor.txt"

func fetchTor() ([]byte, error) {
	a := tor.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncTorData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(torFile, data, wt, fs)
}
