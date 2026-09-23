package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/circleci"
)

const circleciFile = "circleci.json"

func fetchCircleCI() ([]byte, error) {
	a := circleci.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncCircleCIData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(circleciFile, data, wt, fs)
}
