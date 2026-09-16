package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/intercom"
)

const intercomFile = "intercom.json"

func fetchIntercom() ([]byte, error) {
	a := intercom.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncIntercomData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(intercomFile, data, wt, fs)
}
