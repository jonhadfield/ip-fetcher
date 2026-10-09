package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/vultr"
)

const vultrFile = "vultr.json"

func fetchVultr() ([]byte, error) {
	a := vultr.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncVultrData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(vultrFile, data, wt, fs)
}
