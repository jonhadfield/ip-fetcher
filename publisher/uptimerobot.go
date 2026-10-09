package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/uptimerobot"
)

const uptimerobotFile = "uptimerobot.txt"

func fetchUptimerobot() ([]byte, error) {
	a := uptimerobot.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncUptimerobotData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(uptimerobotFile, data, wt, fs)
}
