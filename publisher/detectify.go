package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/detectify"
)

const detectifyFile = "detectify.txt"

func fetchDetectify() ([]byte, error) {
	a := detectify.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncDetectifyData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(detectifyFile, data, wt, fs)
}
