package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/huawei"
)

const huaweiFile = "huawei.json"

func fetchHuawei() ([]byte, error) {
	a := huawei.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncHuaweiData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(huaweiFile, data, wt, fs)
}
