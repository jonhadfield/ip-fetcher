package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/betterstack"
)

const betterstackFile = "betterstack.txt"

func fetchBetterstack() ([]byte, error) {
	a := betterstack.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncBetterstackData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(betterstackFile, data, wt, fs)
}
