package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/spamhaus"
)

const spamhausFile = "spamhaus.json"

func fetchSpamhaus() ([]byte, error) {
	a := spamhaus.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncSpamhausData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(spamhausFile, data, wt, fs)
}
