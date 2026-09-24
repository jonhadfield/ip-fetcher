package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/stopforumspam"
)

const stopforumspamFile = "stopforumspam.txt"

func fetchStopForumSpam() ([]byte, error) {
	a := stopforumspam.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncStopForumSpamData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(stopforumspamFile, data, wt, fs)
}
