package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/intruder"
)

const intruderFile = "intruder.txt"

func fetchIntruder() ([]byte, error) {
	a := intruder.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncIntruderData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(intruderFile, data, wt, fs)
}
