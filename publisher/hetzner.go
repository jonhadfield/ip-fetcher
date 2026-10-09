package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/hetzner"
)

const hetznerFile = "hetzner.json"

func fetchHetzner() ([]byte, error) {
	a := hetzner.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncHetznerData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(hetznerFile, data, wt, fs)
}
