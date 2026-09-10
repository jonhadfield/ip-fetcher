package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/m365"
)

const m365File = "m365.json"

func fetchM365() ([]byte, error) {
	a := m365.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncM365Data(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(m365File, data, wt, fs)
}
