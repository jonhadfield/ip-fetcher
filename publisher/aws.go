package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/aws"
)

const awsFile = "aws.json"

func fetchAWS() ([]byte, error) {
	a := aws.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncAWSData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(awsFile, data, wt, fs)
}
