package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/googleuta"
)

const googleutaFile = "googleuta.json"

func fetchGoogleUTA() ([]byte, error) {
	a := googleuta.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncGoogleUTAData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(googleutaFile, data, wt, fs)
}
