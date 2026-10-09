package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/googleutf"
)

const googleutfFile = "googleutf.json"

func fetchGoogleUTF() ([]byte, error) {
	a := googleutf.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncGoogleUTFData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(googleutfFile, data, wt, fs)
}
