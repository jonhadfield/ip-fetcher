package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/gcp"
)

const gcpFile = "gcp.json"

func fetchGCP() ([]byte, error) {
	a := gcp.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncGCPData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(gcpFile, data, wt, fs)
}
