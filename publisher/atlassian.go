package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/atlassian"
)

const atlassianFile = "atlassian.json"

func fetchAtlassian() ([]byte, error) {
	a := atlassian.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncAtlassianData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(atlassianFile, data, wt, fs)
}
