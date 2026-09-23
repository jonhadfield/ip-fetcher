package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/quiccloud"
)

const quiccloudFile = "quiccloud.txt"

func fetchQuicCloud() ([]byte, error) {
	a := quiccloud.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncQuicCloudData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(quiccloudFile, data, wt, fs)
}
