package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/hetrixtools"
)

const hetrixtoolsFile = "hetrixtools.txt"

func fetchHetrixTools() ([]byte, error) {
	a := hetrixtools.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncHetrixToolsData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(hetrixtoolsFile, data, wt, fs)
}
