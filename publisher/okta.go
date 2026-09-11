package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/okta"
)

const oktaFile = "okta.json"

func fetchOkta() ([]byte, error) {
	a := okta.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncOktaData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(oktaFile, data, wt, fs)
}
