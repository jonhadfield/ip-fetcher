package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/bingbot"
)

const bingbotFile = "bingbot.json"

func fetchBingbot() ([]byte, error) {
	a := bingbot.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncBingbotData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(bingbotFile, data, wt, fs)
}
