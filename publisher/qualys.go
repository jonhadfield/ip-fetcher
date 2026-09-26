package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/qualys"
)

const qualysFile = "qualys.json"

func fetchQualys() ([]byte, error) {
	a := qualys.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncQualysData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(qualysFile, data, wt, fs)
}
