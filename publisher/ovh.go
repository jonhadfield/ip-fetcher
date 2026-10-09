package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/ovh"
)

const ovhFile = "ovh.json"

func fetchOVH() ([]byte, error) {
	a := ovh.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncOVHData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(ovhFile, data, wt, fs)
}
