package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/icloudpr"
)

const icloudprFile = "icloudpr.csv"

func fetchICloudPR() ([]byte, error) {
	a := icloudpr.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncICloudPRData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(icloudprFile, data, wt, fs)
}
