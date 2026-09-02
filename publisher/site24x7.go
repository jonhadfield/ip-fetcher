package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/site24x7"
)

const site24x7File = "site24x7.json"

func fetchSite24x7() ([]byte, error) {
	a := site24x7.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncSite24x7Data(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(site24x7File, data, wt, fs)
}
