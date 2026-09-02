package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/uptrends"
)

const uptrendsFile = "uptrends.json"

func fetchUptrends() ([]byte, error) {
	a := uptrends.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncUptrendsData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(uptrendsFile, data, wt, fs)
}
