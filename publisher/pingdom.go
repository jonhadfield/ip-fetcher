package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/pingdom"
)

const pingdomFile = "pingdom.json"

func fetchPingdom() ([]byte, error) {
	a := pingdom.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncPingdomData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(pingdomFile, data, wt, fs)
}
