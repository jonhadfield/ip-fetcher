package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/m247"
)

const m247File = "m247.json"

func fetchM247() ([]byte, error) {
	a := m247.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncM247Data(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(m247File, data, wt, fs)
}
