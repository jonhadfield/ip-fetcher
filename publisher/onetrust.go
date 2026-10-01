package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/onetrust"
)

const onetrustFile = "onetrust.txt"

func fetchOneTrust() ([]byte, error) {
	a := onetrust.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncOneTrustData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(onetrustFile, data, wt, fs)
}
