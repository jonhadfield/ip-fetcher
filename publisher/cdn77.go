package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/cdn77"
)

const cdn77File = "cdn77.json"

func fetchCDN77() ([]byte, error) {
	a := cdn77.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncCDN77Data(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(cdn77File, data, wt, fs)
}
