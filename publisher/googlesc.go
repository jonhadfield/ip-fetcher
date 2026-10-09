package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/googlesc"
)

const googlescFile = "googlesc.json"

func fetchGoogleSC() ([]byte, error) {
	a := googlesc.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncGoogleSCData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(googlescFile, data, wt, fs)
}
