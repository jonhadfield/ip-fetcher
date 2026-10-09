package publisher

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jonhadfield/ip-fetcher/providers/azure"
)

const azureFile = "azure.json"

func fetchAzure() ([]byte, error) {
	a := azure.New()

	data, _, _, err := a.FetchData()

	return data, err
}

func syncAzureData(data []byte, wt *git.Worktree, fs billy.Filesystem) (plumbing.Hash, error) {
	return syncData(azureFile, data, wt, fs)
}
