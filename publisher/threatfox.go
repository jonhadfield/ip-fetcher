package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/threatfox"
)

const threatfoxFile = "threatfox.json"

func fetchThreatFox() ([]byte, error) {
	a := threatfox.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncThreatFoxData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(threatfoxFile, data, wt, fs)
}
