package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/amazonbot"
)

const amazonbotFile = "amazonbot.json"

func fetchAmazonbot() ([]byte, error) {
	a := amazonbot.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncAmazonbotData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(amazonbotFile, data, wt, fs)
}
