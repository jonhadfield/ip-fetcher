package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/statuscake"
)

const statuscakeFile = "statuscake.json"

func fetchStatuscake() ([]byte, error) {
	a := statuscake.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncStatuscakeData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(statuscakeFile, data, wt, fs)
}
