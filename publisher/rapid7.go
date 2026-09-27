package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/rapid7"
)

const rapid7File = "rapid7.txt"

func fetchRapid7() ([]byte, error) {
	a := rapid7.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncRapid7Data(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(rapid7File, data, wt, fs)
}
