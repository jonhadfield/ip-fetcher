package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/googleutfg"
)

const googleutfgFile = "googleutfg.json"

func fetchGoogleUTFG() ([]byte, error) {
	a := googleutfg.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncGoogleUTFGData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(googleutfgFile, data, wt, fs)
}
