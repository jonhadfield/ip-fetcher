package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/ccbot"
)

const ccbotFile = "ccbot.json"

func fetchCCBot() ([]byte, error) {
	a := ccbot.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncCCBotData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(ccbotFile, data, wt, fs)
}
