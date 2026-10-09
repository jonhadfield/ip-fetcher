package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/alibaba"
)

const alibabaFile = "alibaba.json"

func fetchAlibaba() ([]byte, error) {
	a := alibaba.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncAlibabaData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(alibabaFile, data, wt, fs)
}
