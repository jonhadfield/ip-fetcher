package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/mullvad"
)

const mullvadFile = "mullvad.json"

func fetchMullvad() ([]byte, error) {
	a := mullvad.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncMullvadData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(mullvadFile, data, wt, fs)
}
