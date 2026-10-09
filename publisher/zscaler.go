package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/zscaler"
)

const zscalerFile = "zscaler.json"

func fetchZscaler() ([]byte, error) {
	a := zscaler.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncZscalerData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(zscalerFile, data, wt, fs)
}
