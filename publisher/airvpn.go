package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/airvpn"
)

const airvpnFile = "airvpn.json"

func fetchAirVPN() ([]byte, error) {
	a := airvpn.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncAirVPNData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(airvpnFile, data, wt, fs)
}
