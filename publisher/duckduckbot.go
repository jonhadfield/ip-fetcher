package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/duckduckbot"
)

const duckduckbotFile = "duckduckbot.json"

func fetchDuckduckbot() ([]byte, error) {
	a := duckduckbot.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncDuckduckbotData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(duckduckbotFile, data, wt, fs)
}
