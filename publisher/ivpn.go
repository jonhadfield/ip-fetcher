package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/ivpn"
)

const ivpnFile = "ivpn.json"

func fetchIVPN() ([]byte, error) {
	a := ivpn.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncIVPNData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(ivpnFile, data, wt, fs)
}
