package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/invicti"
)

const invictiFile = "invicti.txt"

func fetchInvicti() ([]byte, error) {
	a := invicti.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncInvictiData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(invictiFile, data, wt, fs)
}
