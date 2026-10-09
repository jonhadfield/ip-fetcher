package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/greensnow"
)

const greensnowFile = "greensnow.txt"

func fetchGreensnow() ([]byte, error) {
	a := greensnow.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncGreensnowData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(greensnowFile, data, wt, fs)
}
