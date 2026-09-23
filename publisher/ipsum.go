package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/ipsum"
)

const ipsumFile = "ipsum.txt"

func fetchIPsum() ([]byte, error) {
	a := ipsum.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncIPsumData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(ipsumFile, data, wt, fs)
}
