package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/leaseweb"
)

const leasewebFile = "leaseweb.json"

func fetchLeaseweb() ([]byte, error) {
	a := leaseweb.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncLeasewebData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(leasewebFile, data, wt, fs)
}
