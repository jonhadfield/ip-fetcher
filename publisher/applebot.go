package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/applebot"
)

const applebotFile = "applebot.json"

func fetchApplebot() ([]byte, error) {
	a := applebot.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncApplebotData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(applebotFile, data, wt, fs)
}
